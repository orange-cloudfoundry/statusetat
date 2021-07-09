package storages

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/orange-cloudfoundry/statusetat/models"
	log "github.com/sirupsen/logrus"
)

func init() {
	gorm.DefaultCallback.Update().Remove("gorm:update_time_stamp")
	gorm.DefaultCallback.Create().Remove("gorm:update_time_stamp")
}

type DB struct {
	db *gorm.DB
}

type Subscriber struct {
	Email string `gorm:"primary_key"`
}

func (s DB) GetDb() *gorm.DB {
	return s.db
}

func (s DB) Creator() func(u *url.URL) (Store, error) {
	return func(u *url.URL) (Store, error) {
		s := &DB{}
		user := ""
		var err error
		if u.User != nil {
			user = u.User.Username()
			password, ok := u.User.Password()
			if ok {
				user += ":" + password
			}
		}
		switch u.Scheme {
		case "mysql":
			fallthrough
		case "mariadb":
			if user != "" {
				user += "@"
			}
			connStr := fmt.Sprintf("%stcp(%s)%s%s", user, u.Host, u.Path, u.RawQuery)
			s.db, err = gorm.Open("mysql", connStr)
			if err != nil {
				return nil, err
			}
		case "sqlite":
			s.db, err = gorm.Open("sqlite3", strings.TrimPrefix(u.String(), "sqlite://"))
			if err != nil {
				return nil, err
			}

		case "postgres":
			s.db, err = gorm.Open("postgres", u.String())
			if err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("sgbd not found")
		}
		if log.IsLevelEnabled(log.DebugLevel) {
			s.db = s.db.Debug()
		}
		s.db.AutoMigrate(&models.Message{}, &models.Incident{}, &models.Metadata{}, &Subscriber{})
		return s, nil
	}
}

func (s DB) Detect(u *url.URL) bool {
	return u.Scheme == "sqlite" || u.Scheme == "mysql" ||
		u.Scheme == "mariadb" || u.Scheme == "postgres"
}

func (s DB) Create(incident models.Incident) (models.Incident, error) {
	for _, msg := range incident.Messages {
		err := s.db.Create(&msg).Error
		if err != nil {
			return incident, err
		}
	}
	err := s.db.Create(&incident).Error
	return incident, err
}

func (s DB) Subscribe(email string) error {
	err := s.db.Create(&Subscriber{Email: email}).Error
	if err != nil {
		return err
	}
	return nil
}

func (s DB) Unsubscribe(email string) error {
	err := s.db.Where("email = ?", email).Delete(Subscriber{}).Error
	if err != nil {
		return err
	}
	return nil
}

func (s DB) Subscribers() ([]string, error) {
	subs := make([]Subscriber, 0)
	err := s.db.Find(&subs).Error
	if err != nil {
		return []string{}, err
	}
	finalSubs := make([]string, len(subs))
	for i, s := range subs {
		finalSubs[i] = s.Email
	}
	return finalSubs, nil
}

func (s DB) Update(guid string, incident models.Incident) (models.Incident, error) {
	err := s.db.Where("incident_guid = ?", guid).Delete(models.Message{}).Error
	if err != nil {
		return incident, err
	}
	err = s.db.Where("incident_guid = ?", guid).Delete(models.Metadata{}).Error
	for i, msg := range incident.Messages {
		err := s.db.Create(&msg).Error
		if err != nil {
			return incident, err
		}
		incident.Messages[i] = msg
	}
	if err != nil {
		return incident, err
	}
	incident.GUID = guid
	var updatedIncident models.Incident
	err = s.db.Model(&updatedIncident).Updates(incident).Error
	if err != nil {
		return updatedIncident, err
	}
	if incident.State == 0 {
		err = s.db.Table("incidents").Where("guid = ?", guid).Update("state", 0).Error
		if err != nil {
			return updatedIncident, err
		}
	}
	if incident.ComponentState == 0 {
		err = s.db.Table("incidents").Where("guid = ?", guid).Update("component_state", 0).Error
		if err != nil {
			return updatedIncident, err
		}
	}

	return updatedIncident, err
}

func (s DB) Delete(guid string) error {
	err := s.db.Where("incident_guid = ?", guid).Delete(models.Message{}).Error
	if err != nil {
		return err
	}
	err = s.db.Where("incident_guid = ?", guid).Delete(models.Metadata{}).Error
	if err != nil {
		return err
	}
	incident := models.Incident{
		GUID: guid,
	}
	err = s.db.Delete(incident).Error
	return err
}

func (s DB) Read(guid string) (models.Incident, error) {
	var incident models.Incident
	err := s.db.Preload("Messages", func(db *gorm.DB) *gorm.DB {
		return db.Order("messages.created_at DESC")
	}).Preload("Metadata").First(&incident, "incidents.guid = ?", guid).Error
	return incident, err
}

func (s DB) ByDate(from, to time.Time) ([]models.Incident, error) {
	var incidents []models.Incident
	err := s.db.Preload("Messages", func(db *gorm.DB) *gorm.DB {
		return db.Order("messages.created_at DESC")
	}).Preload("Metadata").Where("created_at BETWEEN ? AND ?", from, to).Find(&incidents).Error
	return incidents, err
}

func (s DB) Ping() error {
	sdb := s.db.DB()
	return sdb.Ping()
}
