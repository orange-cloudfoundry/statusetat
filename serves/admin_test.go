package serves_test

import (
	"net/http"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/orange-cloudfoundry/statusetat/config"
	"github.com/orange-cloudfoundry/statusetat/models"
)

var _ = Describe("Admin", func() {
	Context("AdminIncidents", func() {
		It("Should give unauthorized when user not set", func() {
			rr := CallRequest(NewRequestInt(http.MethodGet, "/admin/dashboard", nil))
			Expect(rr.CheckError()).To(HaveOccurred())
			Expect(rr.Code).To(Equal(401))
		})
		It("Show only incidents for last 7 days in order", func() {
			cpns := &models.Components{{
				Name:  Component1.Name,
				Group: Component1.Group,
			}}
			inc1 := models.Incident{
				GUID:        "1",
				CreatedAt:   time.Now().AddDate(0, 0, -1).UTC(),
				UpdatedAt:   time.Now().AddDate(0, 0, -1).UTC(),
				Components:  cpns,
				IsScheduled: false,
			}
			inc2 := models.Incident{
				GUID:        "2",
				CreatedAt:   time.Now().UTC(),
				UpdatedAt:   time.Now().UTC(),
				Components:  cpns,
				IsScheduled: false,
			}
			incSched := models.Incident{
				GUID:         "3",
				CreatedAt:    time.Now().AddDate(0, 0, 1).UTC(),
				UpdatedAt:    time.Now().AddDate(0, 0, -1).UTC(),
				Components:   cpns,
				IsScheduled:  true,
				ScheduledEnd: time.Now().AddDate(0, 0, 25).UTC(),
			}
			_, err := fakeStoreMem.Create(inc1)
			Expect(err).ToNot(HaveOccurred())
			_, err = fakeStoreMem.Create(inc2)
			Expect(err).ToNot(HaveOccurred())
			_, err = fakeStoreMem.Create(incSched)
			Expect(err).ToNot(HaveOccurred())

			dataRetrieve := struct {
				BaseInfo  config.BaseInfo
				Incidents []models.Incident
				Before    time.Time
				After     time.Time
			}{}

			fakeHtmlTemplater.ExecuteTemplateStub = TemplateUnmarshalIn("admin/incidents.html", &dataRetrieve)
			rr := CallRequest(NewRequestIntAdmin(http.MethodGet, "/admin/dashboard", nil))
			Expect(rr.CheckError()).ToNot(HaveOccurred())
			Expect(fakeHtmlTemplater.ExecuteTemplateCallCount()).To(Equal(1))
			Expect(dataRetrieve.BaseInfo).To(Equal(BaseInfo))
			Expect(dataRetrieve.Before).ToNot(BeZero())
			Expect(dataRetrieve.After).ToNot(BeZero())
			Expect(dataRetrieve.Incidents).To(HaveLen(2))
			Expect(dataRetrieve.Incidents[0].GUID).To(Equal("2"))
			Expect(dataRetrieve.Incidents[1].GUID).To(Equal("1"))

			CallRequest(NewRequestIntAdmin(http.MethodGet, "/admin/incident", nil))
			Expect(rr.CheckError()).ToNot(HaveOccurred())
			Expect(fakeHtmlTemplater.ExecuteTemplateCallCount()).To(Equal(2))
		})
	})
	Context("AdminMaintenance", func() {
		It("Should give unauthorized when user not set", func() {
			rr := CallRequest(NewRequestInt(http.MethodGet, "/admin/maintenance", nil))
			Expect(rr.CheckError()).To(HaveOccurred())
			Expect(rr.Code).To(Equal(401))
		})
		It("Show only scheduled tasks for next 26 days in order", func() {
			cpns := &models.Components{{
				Name:  Component1.Name,
				Group: Component1.Group,
			}}
			inc1 := models.Incident{
				GUID:        "1",
				CreatedAt:   time.Now().AddDate(0, 0, -1).UTC(),
				UpdatedAt:   time.Now().AddDate(0, 0, -1).UTC(),
				Components:  cpns,
				IsScheduled: false,
			}
			inc2 := models.Incident{
				GUID:        "2",
				CreatedAt:   time.Now().UTC(),
				UpdatedAt:   time.Now().UTC(),
				Components:  cpns,
				IsScheduled: false,
			}
			incSched := models.Incident{
				GUID:         "3",
				CreatedAt:    time.Now().AddDate(0, 0, 25).UTC(),
				UpdatedAt:    time.Now().AddDate(0, 0, 25).UTC(),
				Components:   cpns,
				IsScheduled:  true,
				ScheduledEnd: time.Now().AddDate(0, 0, 25).UTC(),
			}
			_, err := fakeStoreMem.Create(inc1)
			Expect(err).ToNot(HaveOccurred())
			_, err = fakeStoreMem.Create(inc2)
			Expect(err).ToNot(HaveOccurred())
			_, err = fakeStoreMem.Create(incSched)
			Expect(err).ToNot(HaveOccurred())

			dataRetrieve := struct {
				BaseInfo    config.BaseInfo
				Maintenance []models.Incident
			}{}

			fakeHtmlTemplater.ExecuteTemplateStub = TemplateUnmarshalIn("admin/maintenance.html", &dataRetrieve)
			rr := CallRequest(NewRequestIntAdmin(http.MethodGet, "/admin/maintenance", nil))
			Expect(rr.CheckError()).ToNot(HaveOccurred())
			Expect(fakeHtmlTemplater.ExecuteTemplateCallCount()).To(Equal(1))
			Expect(dataRetrieve.BaseInfo).To(Equal(BaseInfo))
			Expect(dataRetrieve.Maintenance).To(HaveLen(1))
			Expect(dataRetrieve.Maintenance[0].GUID).To(Equal("3"))

		})
	})
	Context("AdminAddEditIncidentByType", func() {
		Context("when add", func() {
			Context("Is type incident", func() {
				It("Should give unauthorized when user not set", func() {
					rr := CallRequest(NewRequestInt(http.MethodGet, "/admin/incident/add", nil))
					Expect(rr.CheckError()).To(HaveOccurred())
					Expect(rr.Code).To(Equal(401))
				})
				It("give empty incident with default data", func() {

					dataRetrieve := struct {
						BaseInfo config.BaseInfo
						Incident models.Incident
					}{}

					fakeHtmlTemplater.ExecuteTemplateStub = TemplateUnmarshalIn("admin/add_edit_incident.html", &dataRetrieve)
					rr := CallRequest(NewRequestIntAdmin(http.MethodGet, "/admin/incident/add", nil))
					Expect(rr.CheckError()).ToNot(HaveOccurred())
					Expect(fakeHtmlTemplater.ExecuteTemplateCallCount()).To(Equal(1))
					Expect(dataRetrieve.BaseInfo).To(Equal(BaseInfo))
					Expect(dataRetrieve.Incident.GUID).To(Equal(""))
					Expect(dataRetrieve.Incident.CreatedAt).ToNot(BeZero())
					Expect(dataRetrieve.Incident.ScheduledEnd).ToNot(BeZero())
				})
			})
			Context("Is type scheduled task", func() {
				It("Should give unauthorized when user not set", func() {
					rr := CallRequest(NewRequestInt(http.MethodGet, "/admin/maintenance/add", nil))
					Expect(rr.CheckError()).To(HaveOccurred())
					Expect(rr.Code).To(Equal(401))
				})
				It("give empty scheduled task with default data", func() {

					dataRetrieve := struct {
						BaseInfo config.BaseInfo
						Incident models.Incident
					}{}

					fakeHtmlTemplater.ExecuteTemplateStub = TemplateUnmarshalIn("admin/add_edit_maintenance.html", &dataRetrieve)
					rr := CallRequest(NewRequestIntAdmin(http.MethodGet, "/admin/maintenance/add", nil))
					Expect(rr.CheckError()).ToNot(HaveOccurred())
					Expect(fakeHtmlTemplater.ExecuteTemplateCallCount()).To(Equal(1))
					Expect(dataRetrieve.BaseInfo).To(Equal(BaseInfo))
					Expect(dataRetrieve.Incident.GUID).To(Equal(""))
					Expect(dataRetrieve.Incident.CreatedAt).ToNot(BeZero())
					Expect(dataRetrieve.Incident.ScheduledEnd).ToNot(BeZero())
				})
			})
		})
		Context("when edit", func() {
			BeforeEach(func() {
				cpns := &models.Components{{
					Name:  Component1.Name,
					Group: Component1.Group,
				}}
				inc := models.Incident{
					GUID:        "1",
					CreatedAt:   time.Now().AddDate(0, 0, -1).UTC(),
					UpdatedAt:   time.Now().AddDate(0, 0, -1).UTC(),
					Components:  cpns,
					IsScheduled: false,
				}
				_, err := fakeStoreMem.Create(inc)
				Expect(err).ToNot(HaveOccurred())
				sched := models.Incident{
					GUID:         "2",
					CreatedAt:    time.Now().AddDate(0, 0, -1).UTC(),
					UpdatedAt:    time.Now().AddDate(0, 0, -1).UTC(),
					Components:   cpns,
					IsScheduled:  true,
					ScheduledEnd: time.Now().AddDate(0, 0, -1).UTC(),
				}
				_, err = fakeStoreMem.Create(sched)
				Expect(err).ToNot(HaveOccurred())
			})
			Context("Is type incident", func() {
				It("Should give unauthorized when user not set", func() {
					rr := CallRequest(NewRequestInt(http.MethodGet, "/admin/incident/edit/aguid", nil))
					Expect(rr.CheckError()).To(HaveOccurred())
					Expect(rr.Code).To(Equal(401))
				})
				It("give incident with previous data", func() {

					dataRetrieve := struct {
						BaseInfo config.BaseInfo
						Incident models.Incident
					}{}

					fakeHtmlTemplater.ExecuteTemplateStub = TemplateUnmarshalIn("admin/add_edit_incident.html", &dataRetrieve)
					rr := CallRequest(NewRequestIntAdmin(http.MethodGet, "/admin/incident/edit/1", nil))
					Expect(rr.CheckError()).ToNot(HaveOccurred())
					Expect(fakeHtmlTemplater.ExecuteTemplateCallCount()).To(Equal(1))
					Expect(dataRetrieve.BaseInfo).To(Equal(BaseInfo))
					Expect(dataRetrieve.Incident.GUID).ToNot(BeEmpty())

				})
			})
			Context("Is type scheduled task", func() {
				It("Should give unauthorized when user not set", func() {
					rr := CallRequest(NewRequestInt(http.MethodGet, "/admin/maintenance/edit/aguid", nil))
					Expect(rr.CheckError()).To(HaveOccurred())
					Expect(rr.Code).To(Equal(401))
				})
				It("give scheduled task with previous data", func() {

					dataRetrieve := struct {
						BaseInfo config.BaseInfo
						Incident models.Incident
					}{}

					fakeHtmlTemplater.ExecuteTemplateStub = TemplateUnmarshalIn("admin/add_edit_maintenance.html", &dataRetrieve)
					rr := CallRequest(NewRequestIntAdmin(http.MethodGet, "/admin/maintenance/edit/2", nil))
					Expect(rr.CheckError()).ToNot(HaveOccurred())
					Expect(fakeHtmlTemplater.ExecuteTemplateCallCount()).To(Equal(1))
					Expect(dataRetrieve.BaseInfo).To(Equal(BaseInfo))
					Expect(dataRetrieve.Incident.GUID).ToNot(Equal(""))
					Expect(dataRetrieve.Incident.IsScheduled).To(BeTrue())
					Expect(dataRetrieve.Incident.ScheduledEnd).ToNot(BeZero())
				})
			})
		})
	})
})
