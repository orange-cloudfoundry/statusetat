package storages

import (
	"net/url"
	"os"
	"time"

	"github.com/ArthurHlt/statusetat/models"
)

type Retry struct {
	next    Store
	nbRetry int
}

func NewRetry(next Store, nbRetry int) *Retry {
	return &Retry{next: next, nbRetry: nbRetry}
}

func (m Retry) Creator() func(u *url.URL) (Store, error) {
	return func(u *url.URL) (Store, error) {
		newM := &Retry{}
		store, err := m.next.Creator()(u)
		if err != nil {
			return nil, err
		}
		newM.next = store
		newM.nbRetry = m.nbRetry
		return newM, nil
	}
}

func (m Retry) Detect(u *url.URL) bool {
	return m.next.Detect(u)
}

func (m Retry) Create(incident models.Incident) (models.Incident, error) {
	var err error
	var ret models.Incident
	for i := 0; i < m.nbRetry; i++ {
		ret, err = m.next.Create(incident)
		if err != nil {
			if os.IsNotExist(err) {
				return incident, err
			}
			time.Sleep(500 * time.Millisecond)
			continue
		}
		return ret, err
	}
	return incident, err
}

func (m Retry) Update(guid string, incident models.Incident) (models.Incident, error) {
	var err error
	var ret models.Incident
	for i := 0; i < m.nbRetry; i++ {
		ret, err = m.next.Update(guid, incident)
		if err != nil {
			if os.IsNotExist(err) {
				return incident, err
			}
			time.Sleep(500 * time.Millisecond)
			continue
		}
		return ret, err
	}
	return incident, err
}

func (m Retry) Delete(guid string) error {
	var err error
	for i := 0; i < m.nbRetry; i++ {
		err = m.next.Delete(guid)
		if err != nil {
			if os.IsNotExist(err) {
				return err
			}
			time.Sleep(500 * time.Millisecond)
			continue
		}
		return err
	}
	return err
}

func (m Retry) Read(guid string) (models.Incident, error) {
	var err error
	var ret models.Incident
	for i := 0; i < m.nbRetry; i++ {
		ret, err = m.next.Read(guid)
		if err != nil {
			if os.IsNotExist(err) {
				return models.Incident{}, err
			}
			time.Sleep(500 * time.Millisecond)
			continue
		}
		return ret, err
	}
	return models.Incident{}, err
}

func (m Retry) ByDate(from, to time.Time) ([]models.Incident, error) {
	var err error
	var ret []models.Incident
	for i := 0; i < m.nbRetry; i++ {
		ret, err = m.next.ByDate(from, to)
		if err != nil {
			if os.IsNotExist(err) {
				return []models.Incident{}, err
			}
			time.Sleep(500 * time.Millisecond)
			continue
		}
		return ret, err
	}
	return []models.Incident{}, err
}

func (m Retry) Ping() error {
	var err error
	for i := 0; i < m.nbRetry; i++ {
		err = m.next.Ping()
		if err != nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}
		return err
	}
	return err
}
