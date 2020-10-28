package serves_test

import (
	"net/http"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/orange-cloudfoundry/statusetat/models"
)

var _ = Describe("Cal", func() {
	Context("Ical", func() {
		It("should give back ical data with scheduled task in next 26 days", func() {
			incSched := models.Incident{
				GUID:      "3",
				CreatedAt: time.Now().AddDate(0, 0, 1).UTC(),
				UpdatedAt: time.Now().AddDate(0, 0, -1).UTC(),
				Components: &models.Components{{
					Name:  Component1.Name,
					Group: Component1.Group,
				}},
				Messages: models.Messages{
					{
						CreatedAt: time.Now().AddDate(0, 0, 1).UTC(),
						Title:     "A title",
						Content:   "a content",
					},
				},
				IsScheduled:  true,
				ScheduledEnd: time.Now().AddDate(0, 0, 25).UTC(),
			}
			_, err := fakeStoreMem.Create(incSched)
			Expect(err).ToNot(HaveOccurred())
			rr := CallRequest(NewRequestInt(http.MethodGet, "/cal.ics", nil))
			Expect(rr.CheckError()).ToNot(HaveOccurred())
			Expect(rr.Header().Get("Content-Type")).To(Equal("text/calendar"))
			Expect(rr.Body.String()).To(ContainSubstring("A title"))
			Expect(rr.Body.String()).To(ContainSubstring("a content"))
			Expect(rr.Body.String()).To(ContainSubstring(Component1.Name))

		})
	})
})
