package serves_test

import (
	"net/http"
	"time"

	"github.com/orange-cloudfoundry/statusetat/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Rss", func() {
	Context("rss", func() {
		It("should give back rss with last incidents", func() {
			inc := models.Incident{
				GUID:      "1",
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
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
			}
			_, err := fakeStoreMem.Create(inc)
			Expect(err).ToNot(HaveOccurred())
			rr := CallRequest(NewRequestInt(http.MethodGet, "/rss.xml", nil))
			Expect(rr.CheckError()).ToNot(HaveOccurred())
			Expect(rr.Header().Get("Content-Type")).To(Equal("application/xml"))
			Expect(rr.Body.String()).To(ContainSubstring("A title"))
			Expect(rr.Body.String()).To(ContainSubstring("a content"))
			Expect(rr.Body.String()).To(ContainSubstring(Component1.Name))

		})
	})
	Context("atom", func() {
		It("should give back atom feed with last incidents", func() {
			inc := models.Incident{
				GUID:      "1",
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
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
			}
			_, err := fakeStoreMem.Create(inc)
			Expect(err).ToNot(HaveOccurred())
			rr := CallRequest(NewRequestInt(http.MethodGet, "/atom.xml", nil))
			Expect(rr.CheckError()).ToNot(HaveOccurred())
			Expect(rr.Header().Get("Content-Type")).To(Equal("application/atom+xml"))
			Expect(rr.Body.String()).To(ContainSubstring("A title"))
			Expect(rr.Body.String()).To(ContainSubstring("a content"))
			Expect(rr.Body.String()).To(ContainSubstring(Component1.Name))

		})
	})
})
