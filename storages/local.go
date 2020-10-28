package storages

import (
	"encoding/json"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/orange-cloudfoundry/statusetat/common"
	"github.com/orange-cloudfoundry/statusetat/models"
)

type Local struct {
	dir   string
	mutex *sync.Mutex
}

func (l Local) Creator() func(u *url.URL) (Store, error) {
	return func(u *url.URL) (Store, error) {
		path := strings.TrimPrefix(u.String(), "file://")
		os.MkdirAll(path, 0775)
		return &Local{
			dir:   filepath.FromSlash(strings.TrimSuffix(path, "/")),
			mutex: &sync.Mutex{},
		}, nil
	}
}

func (l Local) Detect(u *url.URL) bool {
	return u.Scheme == "file"
}

func (l Local) Create(incident models.Incident) (models.Incident, error) {
	b, _ := json.Marshal(incident)
	err := ioutil.WriteFile(l.path(incident.GUID), b, 0644)
	return incident, err
}

func (l Local) retrieveSubscribers() ([]string, error) {
	b, err := ioutil.ReadFile(l.path(subscriberFilename))
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return []string{}, err
	}
	subs := make([]string, 0)
	err = json.Unmarshal(b, &subs)
	if err != nil {
		return []string{}, err
	}
	return subs, err
}

func (l Local) storeSubscribers(subscribers []string) error {
	b, _ := json.Marshal(subscribers)
	l.mutex.Lock()
	defer l.mutex.Unlock()
	err := ioutil.WriteFile(l.path(subscriberFilename), b, 0644)
	return err
}

func (l Local) Subscribe(email string) error {
	subs, err := l.retrieveSubscribers()
	if err != nil {
		return err
	}
	if common.InStrSlice(email, subs) {
		return nil
	}
	subs = append(subs, email)
	return l.storeSubscribers(subs)
}

func (l Local) Unsubscribe(email string) error {
	subs, err := l.retrieveSubscribers()
	if err != nil {
		return err
	}
	subs = common.FilterStrSlice(email, subs)
	return l.storeSubscribers(subs)
}

func (l Local) Subscribers() ([]string, error) {
	return l.retrieveSubscribers()
}

func (l Local) Update(guid string, incident models.Incident) (models.Incident, error) {
	b, _ := json.Marshal(incident)
	err := ioutil.WriteFile(l.path(guid), b, 0644)
	return incident, err
}

func (l Local) Delete(guid string) error {
	return os.Remove(l.path(guid))
}

func (l Local) path(fileName string) string {
	return filepath.Join(l.dir, fileName)
}

func (l Local) Read(guid string) (models.Incident, error) {
	var incident models.Incident

	b, err := ioutil.ReadFile(l.path(guid))
	if err != nil {
		return models.Incident{}, err
	}

	err = json.Unmarshal(b, &incident)
	if err != nil {
		return models.Incident{}, err
	}
	return incident, nil
}

func (l Local) ByDate(from, to time.Time) ([]models.Incident, error) {
	incidents := make([]models.Incident, 0)
	err := filepath.Walk(l.dir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		var incident models.Incident

		b, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		err = json.Unmarshal(b, &incident)
		if err != nil {
			return err
		}
		if incident.CreatedAt.After(from) || incident.CreatedAt.Before(to) {
			return nil
		}

		incidents = append(incidents, incident)
		return nil
	})
	return incidents, err
}

func (l Local) Ping() error {
	return nil
}
