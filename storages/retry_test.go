package storages_test

import (
	"fmt"
	"net/url"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/orange-cloudfoundry/statusetat/models"
	"github.com/orange-cloudfoundry/statusetat/storages"
	"github.com/orange-cloudfoundry/statusetat/storages/storagesfakes"
)

var _ = Describe("Retry", func() {
	var store storages.Store
	nbRetry := 3
	var fakeStore *storagesfakes.FakeStore
	BeforeEach(func() {
		fakeStore = &storagesfakes.FakeStore{}
		fakeStore.DetectStub = func(u *url.URL) bool {
			return true
		}
		fakeStore.CreatorStub = func() func(u *url.URL) (storages.Store, error) {
			return func(u *url.URL) (storages.Store, error) {
				return &storagesfakes.FakeStore{}, nil
			}
		}

		store = storages.NewRetryWithSleepTime(fakeStore, nbRetry, 5*time.Millisecond)
	})
	AfterEach(func() {
	})
	Context("Creator", func() {
		It("should create a new instance of retry storage ", func() {
			u, _ := url.Parse("any://any")

			newRetry, err := store.Creator()(u)
			Expect(err).To(BeNil())

			Expect(newRetry).ToNot(BeNil())
		})
	})
	Context("Detect", func() {
	})
	Context("Create", func() {
		It("should retry each time number time defined", func() {
			inc := models.Incident{
				GUID: "aguid",
			}
			fakeStore.CreateStub = func(incident models.Incident) (models.Incident, error) {
				return inc, fmt.Errorf("erroring")
			}
			_, err := store.Create(inc)
			Expect(err).ToNot(BeNil())

			Expect(fakeStore.CreateCallCount()).To(Equal(nbRetry))
		})
		It("should not retry each when next store succeed", func() {
			inc := models.Incident{
				GUID: "aguid",
			}
			fakeStore.CreateStub = func(incident models.Incident) (models.Incident, error) {
				return inc, nil
			}
			_, err := store.Create(inc)
			Expect(err).To(BeNil())

			Expect(fakeStore.CreateCallCount()).To(Equal(1))
		})
	})
	Context("Update", func() {
		It("should retry each time number time defined", func() {
			inc := models.Incident{
				GUID: "aguid",
			}
			fakeStore.UpdateStub = func(guid string, incident models.Incident) (models.Incident, error) {
				return inc, fmt.Errorf("erroring")
			}
			_, err := store.Update(inc.GUID, inc)
			Expect(err).ToNot(BeNil())

			Expect(fakeStore.UpdateCallCount()).To(Equal(nbRetry))
		})
		It("should not retry each when next store succeed", func() {
			inc := models.Incident{
				GUID: "aguid",
			}
			fakeStore.UpdateStub = func(guid string, incident models.Incident) (models.Incident, error) {
				return inc, nil
			}
			_, err := store.Update(inc.GUID, inc)
			Expect(err).To(BeNil())

			Expect(fakeStore.UpdateCallCount()).To(Equal(1))
		})
	})
	Context("Delete", func() {
		It("should retry each time number time defined", func() {
			fakeStore.DeleteStub = func(s string) error {
				return fmt.Errorf("erroring")
			}
			err := store.Delete("aguid")
			Expect(err).ToNot(BeNil())

			Expect(fakeStore.DeleteCallCount()).To(Equal(nbRetry))
		})
		It("should not retry each when next store succeed", func() {
			fakeStore.DeleteStub = func(s string) error {
				return nil
			}
			err := store.Delete("aguid")
			Expect(err).To(BeNil())

			Expect(fakeStore.DeleteCallCount()).To(Equal(1))
		})
	})
	Context("Read", func() {
		It("should retry each time number time defined", func() {
			fakeStore.ReadStub = func(s string) (models.Incident, error) {
				return models.Incident{}, fmt.Errorf("erroring")
			}
			_, err := store.Read("aguid")
			Expect(err).ToNot(BeNil())

			Expect(fakeStore.ReadCallCount()).To(Equal(nbRetry))
		})
		It("should not retry each when next store succeed", func() {
			fakeStore.ReadStub = func(s string) (models.Incident, error) {
				return models.Incident{}, nil
			}
			_, err := store.Read("aguid")
			Expect(err).To(BeNil())

			Expect(fakeStore.ReadCallCount()).To(Equal(1))
		})
	})
	Context("ByDate", func() {
		It("should retry each time number time defined", func() {
			fakeStore.ByDateStub = func(t time.Time, t2 time.Time) ([]models.Incident, error) {
				return []models.Incident{}, fmt.Errorf("erroring")
			}
			_, err := store.ByDate(time.Now(), time.Now())
			Expect(err).ToNot(BeNil())

			Expect(fakeStore.ByDateCallCount()).To(Equal(nbRetry))
		})
		It("should not retry each when next store succeed", func() {
			fakeStore.ByDateStub = func(t time.Time, t2 time.Time) ([]models.Incident, error) {
				return []models.Incident{}, nil
			}
			_, err := store.ByDate(time.Now(), time.Now())
			Expect(err).To(BeNil())

			Expect(fakeStore.ByDateCallCount()).To(Equal(1))
		})
	})

	Context("Subscribe", func() {
		It("should retry each time number time defined", func() {
			fakeStore.SubscribeStub = func(s string) error {
				return fmt.Errorf("erroring")
			}
			err := store.Subscribe("email")
			Expect(err).ToNot(BeNil())

			Expect(fakeStore.SubscribeCallCount()).To(Equal(nbRetry))
		})
		It("should not retry each when next store succeed", func() {
			fakeStore.SubscribeStub = func(s string) error {
				return nil
			}
			err := store.Subscribe("email")
			Expect(err).To(BeNil())

			Expect(fakeStore.SubscribeCallCount()).To(Equal(1))
		})
	})
	Context("Unsubscribe", func() {
		It("should retry each time number time defined", func() {
			fakeStore.UnsubscribeStub = func(s string) error {
				return fmt.Errorf("erroring")
			}
			err := store.Unsubscribe("email")
			Expect(err).ToNot(BeNil())

			Expect(fakeStore.UnsubscribeCallCount()).To(Equal(nbRetry))
		})
		It("should not retry each when next store succeed", func() {
			fakeStore.UnsubscribeStub = func(s string) error {
				return nil
			}
			err := store.Unsubscribe("email")
			Expect(err).To(BeNil())

			Expect(fakeStore.UnsubscribeCallCount()).To(Equal(1))
		})
	})
	Context("Subscribers", func() {
		It("should retry each time number time defined", func() {
			fakeStore.SubscribersStub = func() ([]string, error) {
				return []string{}, fmt.Errorf("erroring")
			}
			_, err := store.Subscribers()
			Expect(err).ToNot(BeNil())

			Expect(fakeStore.SubscribersCallCount()).To(Equal(nbRetry))
		})
		It("should not retry each when next store succeed", func() {
			fakeStore.SubscribersStub = func() ([]string, error) {
				return []string{}, nil
			}
			_, err := store.Subscribers()
			Expect(err).To(BeNil())

			Expect(fakeStore.SubscribersCallCount()).To(Equal(1))
		})
	})
	Context("Ping", func() {
		It("should retry each time number time defined", func() {
			fakeStore.PingStub = func() error {
				return fmt.Errorf("erroring")
			}

			err := store.Ping()
			Expect(err).ToNot(BeNil())

			Expect(fakeStore.PingCallCount()).To(Equal(nbRetry))
		})
		It("should not retry each when next store succeed", func() {
			fakeStore.PingStub = func() error {
				return nil
			}
			err := store.Ping()
			Expect(err).To(BeNil())

			Expect(fakeStore.PingCallCount()).To(Equal(1))
		})
	})
})
