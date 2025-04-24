package storages

import (
	"net/url"
	"os"
	"time"

	"github.com/orange-cloudfoundry/statusetat/v2/models"
)

type Retry struct {
	next      Store
	nbRetry   int
	sleepTime time.Duration
}

func NewRetry(next Store, nbRetry int) *Retry {
	return NewRetryWithSleepTime(next, nbRetry, 500*time.Millisecond)
}

func NewRetryWithSleepTime(next Store, nbRetry int, sleepTime time.Duration) *Retry {
	return &Retry{next: next, nbRetry: nbRetry, sleepTime: sleepTime}
}

func (m *Retry) Creator() func(u *url.URL) (Store, error) {
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

func (m *Retry) Detect(u *url.URL) bool {
	return m.next.Detect(u)
}

func (m *Retry) Create(incident models.Incident) (models.Incident, error) {
	var err error
	var ret models.Incident
	for i := 0; i < m.nbRetry; i++ {
		ret, err = m.next.Create(incident)
		if err != nil {
			if os.IsNotExist(err) {
				return incident, err
			}
			time.Sleep(m.sleepTime)
			continue
		}
		return ret, err
	}
	return incident, err
}

func (m *Retry) Subscribe(email string) error {
	var err error
	for i := 0; i < m.nbRetry; i++ {
		err = m.next.Subscribe(email)
		if err != nil {
			if os.IsNotExist(err) {
				return err
			}
			time.Sleep(m.sleepTime)
			continue
		}
		return err
	}
	return err
}

func (m *Retry) Unsubscribe(email string) error {
	var err error
	for i := 0; i < m.nbRetry; i++ {
		err = m.next.Unsubscribe(email)
		if err != nil {
			if os.IsNotExist(err) {
				return err
			}
			time.Sleep(m.sleepTime)
			continue
		}
		return err
	}
	return err
}

func (m *Retry) Subscribers() ([]string, error) {
	var err error
	var ret []string
	for i := 0; i < m.nbRetry; i++ {
		ret, err = m.next.Subscribers()
		if err != nil {
			if os.IsNotExist(err) {
				return ret, err
			}
			time.Sleep(m.sleepTime)
			continue
		}
		return ret, err
	}
	return []string{}, err
}

func (m *Retry) Update(guid string, incident models.Incident) (models.Incident, error) {
	var err error
	var ret models.Incident
	for i := 0; i < m.nbRetry; i++ {
		ret, err = m.next.Update(guid, incident)
		if err != nil {
			if os.IsNotExist(err) {
				return incident, err
			}
			time.Sleep(m.sleepTime)
			continue
		}
		return ret, err
	}
	return incident, err
}

func (m *Retry) Delete(guid string) error {
	var err error
	for i := 0; i < m.nbRetry; i++ {
		err = m.next.Delete(guid)
		if err != nil {
			if os.IsNotExist(err) {
				return err
			}
			time.Sleep(m.sleepTime)
			continue
		}
		return err
	}
	return err
}

func (m *Retry) Read(guid string) (models.Incident, error) {
	var err error
	var ret models.Incident
	for i := 0; i < m.nbRetry; i++ {
		ret, err = m.next.Read(guid)
		if err != nil {
			if os.IsNotExist(err) {
				return models.Incident{}, err
			}
			time.Sleep(m.sleepTime)
			continue
		}
		return ret, err
	}
	return models.Incident{}, err
}

func (m *Retry) ByDate(from, to time.Time) ([]models.Incident, error) {
	var err error
	var ret []models.Incident
	for i := 0; i < m.nbRetry; i++ {
		ret, err = m.next.ByDate(from, to)
		if err != nil {
			if os.IsNotExist(err) {
				return []models.Incident{}, err
			}
			time.Sleep(m.sleepTime)
			continue
		}
		return ret, err
	}
	return []models.Incident{}, err
}

func (m *Retry) Persistents() ([]models.Incident, error) {
	var err error
	var ret []models.Incident
	for i := 0; i < m.nbRetry; i++ {
		ret, err = m.next.Persistents()
		if err != nil {
			if os.IsNotExist(err) {
				return []models.Incident{}, err
			}
			time.Sleep(m.sleepTime)
			continue
		}
		return ret, err
	}
	return []models.Incident{}, err
}

func (m *Retry) Ping() error {
	var err error
	for i := 0; i < m.nbRetry; i++ {
		err = m.next.Ping()
		if err != nil {
			time.Sleep(m.sleepTime)
			continue
		}
		return err
	}
	return err
}
