package storages_test

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/orange-cloudfoundry/statusetat/models"
	"github.com/orange-cloudfoundry/statusetat/storages"
)

var _ = Describe("Local", func() {
	var localStorage storages.Store
	tmpDirLocal := filepath.Join(os.TempDir(), "statusetat-test-local")
	subscribFile := filepath.Join(tmpDirLocal, "subscribers.json")
	BeforeEach(func() {
		var err error
		u, _ := url.Parse("file://" + tmpDirLocal)
		crea := (&storages.Local{}).Creator()
		localStorage, err = crea(u)
		if err != nil {
			panic(err)
		}
	})
	AfterEach(func() {
		os.RemoveAll(tmpDirLocal)
	})
	Context("Creator", func() {
		It("should create a new instance of local storage ", func() {
			u, _ := url.Parse(fmt.Sprintf("file://%s/my/path", tmpDirLocal))
			defer os.RemoveAll(fmt.Sprintf("file://%s/my/path", tmpDirLocal))

			newLocal, err := localStorage.Creator()(u)
			Expect(err).To(BeNil())

			Expect(newLocal).ToNot(BeNil())
		})
	})
	Context("Detect", func() {
		It("should only detect scheme file://", func() {
			u, _ := url.Parse(fmt.Sprintf("file://%s/my/path", tmpDirLocal))
			defer os.RemoveAll(fmt.Sprintf("file://%s/my/path", tmpDirLocal))
			Expect(localStorage.Detect(u)).To(BeTrue())

			u, _ = url.Parse(fmt.Sprintf("notfile://%s/my/path", tmpDirLocal))
			Expect(localStorage.Detect(u)).To(BeFalse())
		})
	})
	Context("Create", func() {
		It("should create a new json file with incident content", func() {
			inc := models.Incident{
				GUID: "aguid",
			}

			newInc, err := localStorage.Create(inc)
			Expect(newInc).To(BeEquivalentTo(inc))
			Expect(err).To(BeNil())

			Expect(filepath.Join(tmpDirLocal, inc.GUID)).To(BeAnExistingFile())
		})
	})
	Context("Update", func() {
		It("should create a new json file with incident content", func() {
			inc := models.Incident{
				GUID: "aguid",
			}
			newInc, err := localStorage.Create(inc)
			Expect(newInc).To(BeEquivalentTo(inc))
			Expect(err).To(BeNil())
			statFileBefore, err := os.Stat(filepath.Join(tmpDirLocal, inc.GUID))
			Expect(err).To(BeNil())

			inc.Messages = []models.Message{
				{
					GUID:    "aguid-message",
					Title:   "atitle",
					Content: "acontent",
				},
			}
			updatedInc, err := localStorage.Update(inc.GUID, inc)
			Expect(updatedInc).To(BeEquivalentTo(inc))
			Expect(err).To(BeNil())
			statFileUpdated, err := os.Stat(filepath.Join(tmpDirLocal, inc.GUID))
			Expect(err).To(BeNil())

			Expect(statFileUpdated.Size() > statFileBefore.Size()).To(BeTrue(), "file before update is bigger than updated")
		})
	})
	Context("Delete", func() {
		It("should delete file ", func() {
			inc := models.Incident{
				GUID: "aguid",
			}
			newInc, err := localStorage.Create(inc)
			Expect(newInc).To(BeEquivalentTo(inc))
			Expect(err).To(BeNil())
			pathJson := filepath.Join(tmpDirLocal, inc.GUID)
			Expect(pathJson).To(BeAnExistingFile())

			err = localStorage.Delete(inc.GUID)
			Expect(err).To(BeNil())

			Expect(pathJson).ToNot(BeAnExistingFile())
		})
	})
	Context("Read", func() {
		It("should give correct incident", func() {
			inc := models.Incident{
				GUID: "aguid",
			}
			_, err := localStorage.Create(inc)
			Expect(err).To(BeNil())

			newInc, err := localStorage.Read(inc.GUID)
			Expect(err).To(BeNil())

			Expect(newInc).To(BeEquivalentTo(inc))
		})
	})
	Context("ByDate", func() {
		It("Should give incidents in the datetime range", func() {
			d2020 := time.Date(2020, 1, 1, 1, 1, 1, 1, time.Local).UTC()
			d2019 := time.Date(2019, 1, 1, 1, 1, 1, 1, time.Local).UTC()
			d2018 := time.Date(2018, 1, 1, 1, 1, 1, 1, time.Local).UTC()
			inc1 := models.Incident{
				GUID:      "1",
				CreatedAt: d2020,
			}
			inc2 := models.Incident{
				GUID:      "2",
				CreatedAt: d2019,
			}
			inc3 := models.Incident{
				GUID:      "3",
				CreatedAt: d2018,
			}
			_, err := localStorage.Create(inc1)
			Expect(err).To(BeNil())
			_, err = localStorage.Create(inc2)
			Expect(err).To(BeNil())
			_, err = localStorage.Create(inc3)
			Expect(err).To(BeNil())

			incidents, err := localStorage.ByDate(d2019, d2020)
			Expect(err).To(BeNil())
			Expect(incidents).Should(HaveLen(2))
			Expect(incidents[0].CreatedAt).Should(Equal(d2020))
			Expect(incidents[1].CreatedAt).Should(Equal(d2019))

		})
	})

	Context("Subscribe", func() {
		It("should add an user in subscribers.json file", func() {
			Expect(subscribFile).ToNot(BeAnExistingFile())

			err := localStorage.Subscribe("auser")
			Expect(err).To(BeNil())

			Expect(subscribFile).To(BeAnExistingFile())
		})
	})
	Context("Unsubscribe", func() {
		It("should remove entry from subscribers.json file", func() {
			err := localStorage.Subscribe("auser")
			Expect(err).To(BeNil())
			statBefore, err := os.Stat(subscribFile)
			Expect(err).To(BeNil())

			err = localStorage.Unsubscribe("auser")
			Expect(err).To(BeNil())
			statAfter, err := os.Stat(subscribFile)
			Expect(err).To(BeNil())

			Expect(statBefore.Size() > statAfter.Size()).To(BeTrue(), "file before update is bigger than updated")
		})
	})
	Context("Subscribers", func() {
		It("should give all subscribers", func() {
			err := localStorage.Subscribe("auser1")
			Expect(err).To(BeNil())
			err = localStorage.Subscribe("auser2")
			Expect(err).To(BeNil())

			allSubscribers, err := localStorage.Subscribers()
			Expect(err).To(BeNil())
			Expect(allSubscribers).Should(HaveLen(2))
			Expect(allSubscribers[0]).Should(Equal("auser1"))
			Expect(allSubscribers[1]).Should(Equal("auser2"))
		})
	})
	Context("Ping", func() {
		It("should always return nil", func() {
			Expect(localStorage.Ping()).To(BeNil())
		})
	})
})
