package storages

import (
	"encoding/json"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ArthurHlt/statusetat/models"
)

type Local struct {
	dir string
}

func (l Local) Creator() func(u *url.URL) (Store, error) {
	return func(u *url.URL) (Store, error) {
		os.MkdirAll(u.Path, 0775)
		return &Local{
			dir: filepath.FromSlash(strings.TrimSuffix(u.Path, "/")),
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

func (l Local) Update(guid string, incident models.Incident) (models.Incident, error) {
	b, _ := json.Marshal(incident)
	err := ioutil.WriteFile(l.path(guid), b, 0644)
	return incident, err
}

func (l Local) Delete(guid string) error {
	return os.Remove(l.path(guid))
}

func (l Local) path(guid string) string {
	return filepath.Join(l.dir, guid)
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
