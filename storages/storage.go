package storages

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Store

import (
	"fmt"
	"net/url"
	"time"

	"github.com/orange-cloudfoundry/statusetat/models"
)

type Store interface {
	Creator() func(u *url.URL) (Store, error)
	Detect(u *url.URL) bool

	Create(incident models.Incident) (models.Incident, error)
	Update(guid string, incident models.Incident) (models.Incident, error)
	Delete(guid string) error
	Read(guid string) (models.Incident, error)
	ByDate(from, to time.Time) ([]models.Incident, error)

	Subscribe(email string) error
	Unsubscribe(email string) error
	Subscribers() ([]string, error)

	Ping() error
}

var initStores = []Store{
	NewRetry(&DB{}, 3),
	NewRetry(&S3{}, 3),
	NewRetry(&Local{}, 3),
}

func Factory(urls []*url.URL) (Store, error) {
	if len(urls) == 0 {
		return nil, fmt.Errorf("Url store must be set")
	}
	if len(urls) == 1 {
		u := urls[0]
		for _, store := range initStores {
			if !store.Detect(u) {
				continue
			}
			return store.Creator()(u)
		}
		return nil, fmt.Errorf("No valid store can be found")
	}

	repl := NewReplicate(initStores)
	for _, u := range urls {
		_, err := repl.Creator()(u)
		if err != nil {
			return nil, err
		}
	}
	return repl, nil
}
