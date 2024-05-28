package storages

import (
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/orange-cloudfoundry/statusetat/common"
	"github.com/orange-cloudfoundry/statusetat/models"
)

type Local struct {
	dir             string
	mutexSubscriber *sync.Mutex
	mutexPersistent *sync.Mutex
}

func (l Local) Creator() func(u *url.URL) (Store, error) {
	return func(u *url.URL) (Store, error) {
		path := strings.TrimPrefix(u.String(), "file://")
		if err := os.MkdirAll(path, 0775); err != nil {
			return nil, err
		}

		return &Local{
			dir:             filepath.FromSlash(strings.TrimSuffix(path, "/")),
			mutexSubscriber: &sync.Mutex{},
			mutexPersistent: &sync.Mutex{},
		}, nil
	}
}

func (l Local) Detect(u *url.URL) bool {
	return u.Scheme == "file"
}

func (l Local) Create(incident models.Incident) (models.Incident, error) {
	if incident.Persistent {
		err := l.addPersistent(incident)
		return incident, err
	}
	b, _ := json.Marshal(incident)
	err := os.WriteFile(l.path(incident.GUID), b, 0644)
	return incident, err
}

func (l Local) retrieveSubscribers() ([]string, error) {
	b, err := os.ReadFile(l.path(subscriberFilename))
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

func (l Local) Persistents() ([]models.Incident, error) {
	b, err := os.ReadFile(l.path(persistentFilename))
	if err != nil {
		if os.IsNotExist(err) {
			return []models.Incident{}, nil
		}
		return []models.Incident{}, err
	}
	subs := make([]models.Incident, 0)
	err = json.Unmarshal(b, &subs)
	if err != nil {
		return []models.Incident{}, err
	}
	return subs, err
}

func (l Local) addPersistent(incident models.Incident) error {
	incidents, err := l.Persistents()
	if err != nil {
		return err
	}
	incidents = models.Incidents(incidents).Filter(incident.GUID)
	incidents = append(incidents, incident)
	return l.storePersistents(incidents)
}

func (l Local) removePersistent(guid string) error {
	incidents, err := l.Persistents()
	if err != nil {
		return err
	}
	incidents = models.Incidents(incidents).Filter(guid)
	return l.storePersistents(incidents)
}

func (l Local) readPersistent(guid string) (models.Incident, error) {
	incidents, err := l.Persistents()
	if err != nil {
		return models.Incident{}, err
	}
	return models.Incidents(incidents).Find(guid), nil
}

func (l Local) storePersistents(incidents []models.Incident) error {
	sort.Sort(models.Incidents(incidents))
	b, _ := json.Marshal(incidents)
	l.mutexPersistent.Lock()
	defer l.mutexPersistent.Unlock()
	err := os.WriteFile(l.path(persistentFilename), b, 0644)
	return err
}

func (l Local) storeSubscribers(subscribers []string) error {
	b, _ := json.Marshal(subscribers)
	l.mutexSubscriber.Lock()
	defer l.mutexSubscriber.Unlock()
	err := os.WriteFile(l.path(subscriberFilename), b, 0644)
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
	if incident.Persistent {
		_ = l.Delete(guid) // nolint
		err := l.addPersistent(incident)
		return incident, err
	}

	_ = l.removePersistent(guid) // nolint
	b, _ := json.Marshal(incident)
	err := os.WriteFile(l.path(guid), b, 0644)
	return incident, err
}

func (l Local) Delete(guid string) error {
	err := l.removePersistent(guid)
	if err != nil {
		return err
	}
	return os.Remove(l.path(guid))
}

func (l Local) path(fileName string) string {
	return filepath.Join(l.dir, fileName)
}

func (l Local) Read(guid string) (models.Incident, error) {
	incident, err := l.readPersistent(guid)
	if err != nil {
		return models.Incident{}, err
	}
	if incident.GUID == guid {
		return incident, nil
	}

	b, err := os.ReadFile(l.path(guid))
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
		if filepath.Base(path) == subscriberFilename ||
			filepath.Base(path) == persistentFilename {
			return nil
		}
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		var incident models.Incident

		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		err = json.Unmarshal(b, &incident)
		if err != nil {
			return err
		}
		if incident.CreatedAt.Before(from) || incident.CreatedAt.After(to) {
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
