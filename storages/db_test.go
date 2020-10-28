package storages_test

import (
	"net/url"
	"time"

	"github.com/jinzhu/gorm"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/orange-cloudfoundry/statusetat/models"
	"github.com/orange-cloudfoundry/statusetat/storages"
)

var _ = Describe("Db", func() {
	var store storages.Store
	var db *gorm.DB
	BeforeEach(func() {
		var err error
		u, _ := url.Parse("sqlite://:memory:")
		crea := (&storages.DB{}).Creator()
		store, err = crea(u)
		if err != nil {
			panic(err)
		}
		db = store.(*storages.DB).GetDb()
	})
	Context("Creator", func() {
		It("should create a new instance of local storage ", func() {
			u, _ := url.Parse("sqlite://:memory:")

			newLocal, err := store.Creator()(u)
			Expect(err).To(BeNil())

			Expect(newLocal).ToNot(BeNil())
		})
	})
	Context("Detect", func() {
		It("should detect scheme sqlite or mysql or mariadb or postgres", func() {
			u, _ := url.Parse("sqlite://toto.com")
			Expect(store.Detect(u)).To(BeTrue())

			u, _ = url.Parse("mysql://toto.com")
			Expect(store.Detect(u)).To(BeTrue())

			u, _ = url.Parse("mariadb://toto.com")
			Expect(store.Detect(u)).To(BeTrue())

			u, _ = url.Parse("postgres://toto.com")
			Expect(store.Detect(u)).To(BeTrue())
		})
	})
	Context("Create", func() {
		It("should add incident to db", func() {
			inc := models.Incident{
				GUID:      "aguid",
				CreatedAt: time.Now().In(time.UTC),
				UpdatedAt: time.Now().In(time.UTC),
				Messages:  []models.Message{},
				Metadata:  []models.Metadata{},
			}

			newInc, err := store.Create(inc)
			Expect(newInc).To(BeEquivalentTo(inc))
			Expect(err).To(BeNil())

			var incDb models.Incident
			db.First(&incDb)

			Expect(incDb.GUID).To(Equal(inc.GUID))
		})
	})
	Context("Update", func() {
		It("should update incident in db", func() {
			inc := models.Incident{
				GUID:      "aguid",
				CreatedAt: time.Now().In(time.UTC),
				UpdatedAt: time.Now().In(time.UTC),
				Messages:  []models.Message{},
				Metadata:  []models.Metadata{},
			}
			_, err := store.Create(inc)
			Expect(err).To(BeNil())
			inc.Messages = []models.Message{
				{
					GUID:      "aguid-message",
					Title:     "atitle",
					CreatedAt: time.Now().In(time.UTC),
					Content:   "acontent",
				},
			}
			_, err = store.Update(inc.GUID, inc)
			Expect(err).To(BeNil())

			var incDb models.Incident
			db.Preload("Messages", func(db *gorm.DB) *gorm.DB {
				return db.Order("messages.created_at DESC")
			}).Preload("Metadata").First(&incDb)

			Expect(incDb).To(BeEquivalentTo(inc))
		})
	})
	Context("Delete", func() {
		It("should delete entry in db ", func() {
			inc := models.Incident{
				GUID:      "aguid",
				CreatedAt: time.Now().In(time.UTC),
				UpdatedAt: time.Now().In(time.UTC),
				Messages:  []models.Message{},
				Metadata:  []models.Metadata{},
			}
			newInc, err := store.Create(inc)
			Expect(newInc).To(BeEquivalentTo(inc))
			Expect(err).To(BeNil())

			err = store.Delete(inc.GUID)
			Expect(err).To(BeNil())

			incDb := &models.Incident{}
			db.First(incDb)

			Expect(incDb.GUID).To(Equal(""))
		})
	})
	Context("Read", func() {
		It("should give correct incident", func() {
			inc := models.Incident{
				GUID:      "aguid",
				CreatedAt: time.Now().In(time.UTC),
				UpdatedAt: time.Now().In(time.UTC),
				Messages:  []models.Message{},
				Metadata:  []models.Metadata{},
			}
			_, err := store.Create(inc)
			Expect(err).To(BeNil())

			newInc, err := store.Read(inc.GUID)
			Expect(err).To(BeNil())

			Expect(newInc.CreatedAt).To(Equal(inc.CreatedAt))
			Expect(newInc).To(BeEquivalentTo(inc))
		})
	})
	Context("ByDate", func() {
		It("Should give incidents in the datetime range", func() {
			d2020 := time.Date(2020, 1, 1, 1, 1, 1, 0, time.UTC)
			d2019 := time.Date(2019, 1, 1, 1, 1, 1, 0, time.UTC)
			d2018 := time.Date(2018, 1, 1, 1, 1, 1, 0, time.UTC)
			inc1 := models.Incident{
				GUID:      "1",
				CreatedAt: d2020,
				UpdatedAt: d2020,
			}
			inc2 := models.Incident{
				GUID:      "2",
				CreatedAt: d2019,
				UpdatedAt: d2019,
			}
			inc3 := models.Incident{
				GUID:      "3",
				CreatedAt: d2018,
				UpdatedAt: d2018,
			}
			_, err := store.Create(inc1)
			Expect(err).To(BeNil())
			_, err = store.Create(inc2)
			Expect(err).To(BeNil())
			_, err = store.Create(inc3)
			Expect(err).To(BeNil())

			incidents, err := store.ByDate(d2020, d2019)
			Expect(err).To(BeNil())
			Expect(incidents).Should(HaveLen(2))

			Expect(incidents[0].CreatedAt).Should(Equal(d2020))
			Expect(incidents[1].CreatedAt).Should(Equal(d2019))
		})
	})

	Context("Subscribe", func() {
		It("should add an user db", func() {
			err := store.Subscribe("auser")
			Expect(err).To(BeNil())

			subscribers := make([]storages.Subscriber, 0)
			db.Find(&subscribers)
			Expect(subscribers).To(HaveLen(1))
			Expect(subscribers[0].Email).To(Equal("auser"))

		})
	})
	Context("Unsubscribe", func() {
		It("should remove entry from subscribers.json file", func() {
			err := store.Subscribe("auser")
			Expect(err).To(BeNil())

			err = store.Unsubscribe("auser")
			Expect(err).To(BeNil())

			subscribers := make([]storages.Subscriber, 0)
			db.Find(&subscribers)
			Expect(subscribers).To(BeEmpty())
		})
	})
	Context("Subscribers", func() {
		It("should give all subscribers", func() {
			err := store.Subscribe("auser1")
			Expect(err).To(BeNil())
			err = store.Subscribe("auser2")
			Expect(err).To(BeNil())

			allSubscribers, err := store.Subscribers()
			Expect(err).To(BeNil())
			Expect(allSubscribers).Should(HaveLen(2))
			Expect(allSubscribers[0]).Should(Equal("auser1"))
			Expect(allSubscribers[1]).Should(Equal("auser2"))
		})
	})
	Context("Ping", func() {
		It("should always return nil", func() {
			Expect(store.Ping()).To(BeNil())
		})
	})
})
