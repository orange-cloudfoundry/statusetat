package storages_test

import (
	"fmt"
	"net/url"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/orange-cloudfoundry/statusetat/v2/models"
	"github.com/orange-cloudfoundry/statusetat/v2/storages"
	"github.com/orange-cloudfoundry/statusetat/v2/storages/storagesfakes"
)

var _ = Describe("Replicate", func() {
	var store storages.Store
	var fakeStore1 *storagesfakes.FakeStore
	var fakeStore2 *storagesfakes.FakeStore
	BeforeEach(func() {
		fakeStore1 = &storagesfakes.FakeStore{}
		fakeStore1.DetectStub = func(u *url.URL) bool {
			return u.Scheme == "fake1"
		}
		fakeStore1.CreatorStub = func() func(u *url.URL) (storages.Store, error) {
			return func(u *url.URL) (storages.Store, error) {
				return fakeStore1, nil
			}
		}

		fakeStore2 = &storagesfakes.FakeStore{}
		fakeStore2.DetectStub = func(u *url.URL) bool {
			return u.Scheme == "fake2"
		}
		fakeStore2.CreatorStub = func() func(u *url.URL) (storages.Store, error) {
			return func(u *url.URL) (storages.Store, error) {
				return fakeStore2, nil
			}
		}
		u1, _ := url.Parse("fake1:///any")
		u2, _ := url.Parse("fake2:///any")
		store = storages.NewReplicateWithWaits([]storages.Store{fakeStore1, fakeStore2}, 5*time.Millisecond, 1*time.Hour)
		_, err := store.Creator()(u1)
		if err != nil {
			panic(err)
		}
		_, err = store.Creator()(u2)
		if err != nil {
			panic(err)
		}
	})
	AfterEach(func() {
	})
	Context("Creator", func() {
		It("should create a new instance of replicate storage ", func() {
			u, _ := url.Parse("fake1://any")

			newReplicate, err := store.Creator()(u)
			Expect(err).To(BeNil())

			Expect(newReplicate).ToNot(BeNil())
		})
	})
	Context("Detect", func() {
	})
	Context("Create", func() {
		It("should replay after time when erroring on one store", func() {
			inc := models.Incident{
				GUID: "aguid",
			}

			fakeStore1.CreateStub = func(incident models.Incident) (models.Incident, error) {
				return inc, nil
			}

			fakeStore2.CreateStub = func(incident models.Incident) (models.Incident, error) {
				if fakeStore2.CreateCallCount() == 2 {
					return inc, nil
				}
				return inc, fmt.Errorf("erroring")
			}
			_, err := store.Create(inc)
			Expect(err).To(BeNil())

			Expect(fakeStore1.CreateCallCount()).To(Equal(1))
			Expect(fakeStore2.CreateCallCount()).To(Equal(1))
			time.Sleep(6 * time.Millisecond)
			Expect(fakeStore1.CreateCallCount()).To(Equal(1))
			Expect(fakeStore2.CreateCallCount()).To(Equal(2))
		})
	})
	Context("Update", func() {
		It("should replay after time when erroring on one store", func() {
			inc := models.Incident{
				GUID: "aguid",
			}

			fakeStore1.UpdateStub = func(guid string, incident models.Incident) (models.Incident, error) {
				return inc, nil
			}

			fakeStore2.UpdateStub = func(guid string, incident models.Incident) (models.Incident, error) {
				if fakeStore2.UpdateCallCount() == 2 {
					return inc, nil
				}
				return inc, fmt.Errorf("erroring")
			}
			_, err := store.Update("aguid", inc)
			Expect(err).To(BeNil())

			Expect(fakeStore1.UpdateCallCount()).To(Equal(1))
			Expect(fakeStore2.UpdateCallCount()).To(Equal(1))
			time.Sleep(6 * time.Millisecond)
			Expect(fakeStore1.UpdateCallCount()).To(Equal(1))
			Expect(fakeStore2.UpdateCallCount()).To(Equal(2))
		})
	})
	Context("Delete", func() {
		It("should replay after time when erroring on one store", func() {
			fakeStore1.DeleteStub = func(guid string) error {
				return nil
			}

			fakeStore2.DeleteStub = func(guid string) error {
				if fakeStore2.DeleteCallCount() == 2 {
					return nil
				}
				return fmt.Errorf("erroring")
			}

			err := store.Delete("aguid")
			Expect(err).To(BeNil())

			Expect(fakeStore1.DeleteCallCount()).To(Equal(1))
			Expect(fakeStore2.DeleteCallCount()).To(Equal(1))
			time.Sleep(6 * time.Millisecond)
			Expect(fakeStore1.DeleteCallCount()).To(Equal(1))
			Expect(fakeStore2.DeleteCallCount()).To(Equal(2))
		})
	})
	Context("Read", func() {
		It("should take information from first responding without error store", func() {
			fakeStore1.ReadStub = func(s string) (models.Incident, error) {
				return models.Incident{
					GUID: "guid-fake1",
				}, fmt.Errorf("erroring")
			}
			fakeStore2.ReadStub = func(s string) (models.Incident, error) {
				return models.Incident{
					GUID: "guid-fake2",
				}, nil
			}

			inc, err := store.Read("aguid")
			Expect(err).To(BeNil())

			Expect(fakeStore2.ReadCallCount()).To(Equal(1))
			Expect(inc.GUID).To(Equal("guid-fake2"))
		})
	})
	Context("ByDate", func() {
		It("should take information from first responding without error store", func() {
			fakeStore1.ByDateStub = func(t time.Time, t2 time.Time) ([]models.Incident, error) {
				return []models.Incident{}, fmt.Errorf("erroring")
			}
			fakeStore2.ByDateStub = func(t time.Time, t2 time.Time) ([]models.Incident, error) {
				return []models.Incident{
					{
						GUID: "guid-fake2",
					},
				}, nil
			}

			incs, err := store.ByDate(time.Now(), time.Now())
			Expect(err).To(BeNil())

			Expect(fakeStore2.ByDateCallCount()).To(Equal(1))
			Expect(incs).NotTo(BeEmpty())
			Expect(incs[0].GUID).To(Equal("guid-fake2"))
		})
	})

	Context("Persistents", func() {
		It("should take information from first responding without error store", func() {
			fakeStore1.PersistentsStub = func() ([]models.Incident, error) {
				return []models.Incident{}, fmt.Errorf("erroring")
			}
			fakeStore2.PersistentsStub = func() ([]models.Incident, error) {
				return []models.Incident{
					{
						GUID: "guid-fake2",
					},
				}, nil
			}

			incs, err := store.Persistents()
			Expect(err).To(BeNil())

			Expect(fakeStore2.PersistentsCallCount()).To(Equal(1))
			Expect(incs).NotTo(BeEmpty())
			Expect(incs[0].GUID).To(Equal("guid-fake2"))
		})
	})

	Context("Subscribe", func() {
		It("should replay after time when erroring on one store", func() {
			fakeStore1.SubscribeStub = func(email string) error {
				return nil
			}

			fakeStore2.SubscribeStub = func(email string) error {
				if fakeStore2.SubscribeCallCount() == 2 {
					return nil
				}
				return fmt.Errorf("erroring")
			}

			err := store.Subscribe("email")
			Expect(err).To(BeNil())

			Expect(fakeStore1.SubscribeCallCount()).To(Equal(1))
			Expect(fakeStore2.SubscribeCallCount()).To(Equal(1))
			time.Sleep(6 * time.Millisecond)
			Expect(fakeStore1.SubscribeCallCount()).To(Equal(1))
			Expect(fakeStore2.SubscribeCallCount()).To(Equal(2))
		})
	})
	Context("Unsubscribe", func() {
		It("should replay after time when erroring on one store", func() {
			fakeStore1.UnsubscribeStub = func(email string) error {
				return nil
			}

			fakeStore2.UnsubscribeStub = func(email string) error {
				if fakeStore2.UnsubscribeCallCount() == 2 {
					return nil
				}
				return fmt.Errorf("erroring")
			}

			err := store.Unsubscribe("email")
			Expect(err).To(BeNil())

			Expect(fakeStore1.UnsubscribeCallCount()).To(Equal(1))
			Expect(fakeStore2.UnsubscribeCallCount()).To(Equal(1))
			time.Sleep(6 * time.Millisecond)
			Expect(fakeStore1.UnsubscribeCallCount()).To(Equal(1))
			Expect(fakeStore2.UnsubscribeCallCount()).To(Equal(2))
		})
	})
	Context("Subscribers", func() {
		It("should take information from first responding without error store", func() {
			fakeStore1.SubscribersStub = func() ([]string, error) {
				return []string{}, fmt.Errorf("erroring")
			}
			fakeStore2.SubscribersStub = func() ([]string, error) {
				return []string{"data"}, nil
			}

			subs, err := store.Subscribers()
			Expect(err).To(BeNil())

			Expect(fakeStore2.SubscribersCallCount()).To(Equal(1))
			Expect(subs).To(HaveLen(1))
			Expect(subs[0]).To(Equal("data"))
		})
	})
	Context("Ping", func() {
	})
})
