package serves_test

import (
	"fmt"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/orange-cloudfoundry/statusetat/models"
	"github.com/orange-cloudfoundry/statusetat/serves"
)

var _ = Describe("Api", func() {
	Context("CreateIncident", func() {
		It("Should give unauthorized when user not set", func() {
			rr := CallRequest(NewRequestInt(http.MethodPost, "/v1/incidents", nil))
			Expect(rr.CheckError()).To(HaveOccurred())
			Expect(rr.Code).To(Equal(401))
		})
		It("should force user to set a first message", func() {
			inc := models.Incident{
				Components: &models.Components{{
					Name:  Component1.Name,
					Group: Component1.Group,
				}},
			}
			rr := CallRequest(NewRequestIntAdmin(http.MethodPost, "/v1/incidents", inc))
			err := rr.CheckError()
			Expect(err).To(HaveOccurred())
			httpErr, ok := err.(serves.HttpError)
			Expect(ok).To(BeTrue(), "is not an http error")
			httpErr.Status = http.StatusPreconditionFailed
			httpErr.Description = "At least one message must be set"
		})
		It("should give back setted incident, store it and emit in emitter an incident data", func() {

			inc := models.Incident{
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
			rr := CallRequest(NewRequestIntAdmin(http.MethodPost, "/v1/incidents", inc))
			Expect(rr.CheckError()).ToNot(HaveOccurred())
			finalInc, err := rr.UnmarshalToIncident()
			Expect(err).ToNot(HaveOccurred())
			Expect(finalInc.GUID).ToNot(BeEmpty())
			Expect(finalInc.Origin).ToNot(BeEmpty())
			Expect(finalInc.CreatedAt).ToNot(BeZero())
			Expect(finalInc.UpdatedAt).ToNot(BeZero())

			Expect(fakeEmitter.EmitCallCount()).To(Equal(1))
			Expect(fakeStoreMem.CreateCallCount()).To(Equal(1))

			dbInc, err := fakeStoreMem.Read(finalInc.GUID)
			Expect(err).ToNot(HaveOccurred())
			Expect(dbInc.GUID).ToNot(BeEmpty())
		})
		Context("is a scheduled task", func() {
			It("should give an error if scheduled end not set", func() {

				inc := models.Incident{
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
					IsScheduled: true,
				}
				rr := CallRequest(NewRequestIntAdmin(http.MethodPost, "/v1/incidents", inc))
				err := rr.CheckError()
				Expect(err).To(HaveOccurred())
				httpErr, ok := err.(serves.HttpError)
				Expect(ok).To(BeTrue(), "is not an http error")
				httpErr.Status = http.StatusPreconditionFailed
				httpErr.Description = "Start date of scheduled maintenance can't be before end date"
			})
		})
	})
	Context("ByDate", func() {
		BeforeEach(func() {
			cpns := &models.Components{{
				Name:  Component1.Name,
				Group: Component1.Group,
			}}
			inc1 := models.Incident{
				GUID:        "1",
				CreatedAt:   time.Now().AddDate(0, 0, -2).UTC(),
				UpdatedAt:   time.Now().AddDate(0, 0, -2).UTC(),
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
			inc4 := models.Incident{
				GUID:        "4",
				CreatedAt:   time.Now().AddDate(0, 0, -8).UTC(),
				UpdatedAt:   time.Now().AddDate(0, 0, -8).UTC(),
				Components:  cpns,
				IsScheduled: false,
			}
			_, err := fakeStoreMem.Create(inc1)
			Expect(err).ToNot(HaveOccurred())
			_, err = fakeStoreMem.Create(inc2)
			Expect(err).ToNot(HaveOccurred())
			_, err = fakeStoreMem.Create(incSched)
			Expect(err).ToNot(HaveOccurred())
			_, err = fakeStoreMem.Create(inc4)
			Expect(err).ToNot(HaveOccurred())
		})
		Context("without all_types param set", func() {
			It("should give ordered incidents from the last 7 days when user not set date", func() {

				rr := CallRequest(NewRequestInt(http.MethodGet, "/v1/incidents", nil))
				Expect(rr.CheckError()).ToNot(HaveOccurred())

				finalIncidents, err := rr.UnmarshalToIncidents()
				Expect(err).ToNot(HaveOccurred())
				Expect(finalIncidents).To(HaveLen(2))
				Expect(finalIncidents[0].GUID).To(Equal("2"))
				Expect(finalIncidents[1].GUID).To(Equal("1"))
			})
			It("should give ordered incidents between date given by user", func() {
				from := time.Now().AddDate(0, 0, -10).UTC()
				to := time.Now().AddDate(0, 0, 2).UTC()
				target := fmt.Sprintf(
					"/v1/incidents?from=%s&to=%s",
					from.Format(time.RFC3339),
					to.Format(time.RFC3339),
				)
				rr := CallRequest(NewRequestInt(http.MethodGet, target, nil))
				Expect(rr.CheckError()).ToNot(HaveOccurred())

				finalIncidents, err := rr.UnmarshalToIncidents()
				Expect(err).ToNot(HaveOccurred())
				Expect(finalIncidents).To(HaveLen(3))
				Expect(finalIncidents[0].GUID).To(Equal("2"))
				Expect(finalIncidents[1].GUID).To(Equal("1"))
				Expect(finalIncidents[2].GUID).To(Equal("4"))
			})

		})
		Context("with all_types param set", func() {
			It("should give ordered incidents between date given by user", func() {
				from := time.Now().AddDate(0, 0, -10).UTC()
				to := time.Now().AddDate(0, 0, 2).UTC()
				target := fmt.Sprintf(
					"/v1/incidents?from=%s&to=%s&all_types=true",
					from.Format(time.RFC3339),
					to.Format(time.RFC3339),
				)
				rr := CallRequest(NewRequestInt(http.MethodGet, target, nil))
				Expect(rr.CheckError()).ToNot(HaveOccurred())

				finalIncidents, err := rr.UnmarshalToIncidents()
				Expect(err).ToNot(HaveOccurred())
				Expect(finalIncidents).To(HaveLen(4))
				Expect(finalIncidents[0].GUID).To(Equal("3"))
				Expect(finalIncidents[0].IsScheduled).To(BeTrue())
				Expect(finalIncidents[1].GUID).To(Equal("2"))
				Expect(finalIncidents[2].GUID).To(Equal("1"))
				Expect(finalIncidents[3].GUID).To(Equal("4"))
			})

		})

	})
	Context("Incident", func() {
		It("should give incident found", func() {
			cpns := &models.Components{{
				Name:  Component1.Name,
				Group: Component1.Group,
			}}
			inc1 := models.Incident{
				GUID:        "1",
				CreatedAt:   time.Now().AddDate(0, 0, -2).UTC(),
				UpdatedAt:   time.Now().AddDate(0, 0, -2).UTC(),
				Components:  cpns,
				IsScheduled: false,
			}
			_, err := fakeStoreMem.Create(inc1)
			Expect(err).ToNot(HaveOccurred())

			rr := CallRequest(NewRequestInt(http.MethodGet, "/v1/incidents/1", nil))
			Expect(rr.CheckError()).ToNot(HaveOccurred())

			finalIncident, err := rr.UnmarshalToIncident()
			Expect(err).ToNot(HaveOccurred())

			Expect(finalIncident.GUID).To(Equal("1"))
		})

	})
	Context("Delete", func() {
		It("Should give unauthorized when user not set", func() {
			rr := CallRequest(NewRequestInt(http.MethodDelete, "/v1/incidents/1", nil))
			Expect(rr.CheckError()).To(HaveOccurred())
			Expect(rr.Code).To(Equal(401))
		})
		It("should give incident found", func() {
			cpns := &models.Components{{
				Name:  Component1.Name,
				Group: Component1.Group,
			}}
			inc1 := models.Incident{
				GUID:        "1",
				CreatedAt:   time.Now().AddDate(0, 0, -2).UTC(),
				UpdatedAt:   time.Now().AddDate(0, 0, -2).UTC(),
				Components:  cpns,
				IsScheduled: false,
			}
			_, err := fakeStoreMem.Create(inc1)
			Expect(err).ToNot(HaveOccurred())

			rr := CallRequest(NewRequestIntAdmin(http.MethodDelete, "/v1/incidents/1", nil))
			Expect(rr.CheckError()).ToNot(HaveOccurred())

			_, err = fakeStoreMem.Read("1")
			Expect(err).To(HaveOccurred())
		})

	})
	Context("Update", func() {
		It("Should give unauthorized when user not set", func() {
			rr := CallRequest(NewRequestInt(http.MethodPut, "/v1/incidents/1", nil))
			Expect(rr.CheckError()).To(HaveOccurred())
			Expect(rr.Code).To(Equal(401))
		})
		It("should only update modified field and emit new updated incident", func() {
			cpns := &models.Components{{
				Name:  Component1.Name,
				Group: Component1.Group,
			}}
			updateBefore := time.Now().AddDate(0, 0, -2).UTC()
			inc1 := models.Incident{
				GUID:        "1",
				CreatedAt:   time.Now().AddDate(0, 0, -2).UTC(),
				UpdatedAt:   updateBefore,
				Components:  cpns,
				IsScheduled: false,
				State:       models.Monitoring,
			}
			_, err := fakeStoreMem.Create(inc1)
			Expect(err).ToNot(HaveOccurred())

			rr := CallRequest(NewRequestIntAdmin(http.MethodPut, "/v1/incidents/1", models.Incident{
				State: models.Unresolved,
			}))
			Expect(rr.CheckError()).ToNot(HaveOccurred())
			Expect(fakeEmitter.EmitCallCount()).To(Equal(1))

			finalIncident, err := fakeStoreMem.Read("1")
			Expect(err).ToNot(HaveOccurred())

			Expect(finalIncident.GUID).To(Equal("1"))
			Expect(finalIncident.UpdatedAt).ToNot(Equal(updateBefore))
			Expect(finalIncident.State).To(Equal(models.Unresolved))
			Expect(*finalIncident.Components).To(HaveLen(1))
			Expect((*finalIncident.Components)[0].Name).To(Equal(Component1.Name))
			Expect((*finalIncident.Components)[0].Group).To(Equal(Component1.Group))

		})
		Context("when no_notify is set to true", func() {

			It("should only update modified field and do not emit updated incident", func() {
				cpns := &models.Components{{
					Name:  Component1.Name,
					Group: Component1.Group,
				}}
				updateBefore := time.Now().AddDate(0, 0, -2).UTC()
				inc1 := models.Incident{
					GUID:        "1",
					CreatedAt:   time.Now().AddDate(0, 0, -2).UTC(),
					UpdatedAt:   updateBefore,
					Components:  cpns,
					IsScheduled: false,
					State:       models.Monitoring,
				}
				_, err := fakeStoreMem.Create(inc1)
				Expect(err).ToNot(HaveOccurred())
				state := models.Unresolved
				rr := CallRequest(NewRequestIntAdmin(http.MethodPut, "/v1/incidents/1", models.IncidentUpdateRequest{
					State:    &state,
					NoNotify: true,
				}))
				Expect(rr.CheckError()).ToNot(HaveOccurred())
				Expect(fakeEmitter.EmitCallCount()).To(Equal(0))

				finalIncident, err := fakeStoreMem.Read("1")
				Expect(err).ToNot(HaveOccurred())

				Expect(finalIncident.GUID).To(Equal("1"))
				Expect(finalIncident.UpdatedAt).ToNot(Equal(updateBefore))
				Expect(finalIncident.State).To(Equal(models.Unresolved))
				Expect(*finalIncident.Components).To(HaveLen(1))
				Expect((*finalIncident.Components)[0].Name).To(Equal(Component1.Name))
				Expect((*finalIncident.Components)[0].Group).To(Equal(Component1.Group))

			})
		})
		It("should override all message when param partial_update_message not set", func() {
			cpns := &models.Components{{
				Name:  Component1.Name,
				Group: Component1.Group,
			}}
			inc1 := models.Incident{
				GUID:       "1",
				CreatedAt:  time.Now().AddDate(0, 0, -2).UTC(),
				UpdatedAt:  time.Now().AddDate(0, 0, -2).UTC(),
				Components: cpns,
				Messages: models.Messages{
					{
						GUID:         "1",
						IncidentGUID: "1",
						Title:        "a title",
						Content:      "a content",
					},
					{
						GUID:         "2",
						IncidentGUID: "1",
						Title:        "a sub title",
						Content:      "a sub content",
					},
				},
				IsScheduled: false,
				State:       models.Monitoring,
			}
			_, err := fakeStoreMem.Create(inc1)
			Expect(err).ToNot(HaveOccurred())

			rr := CallRequest(NewRequestIntAdmin(http.MethodPut, "/v1/incidents/1", models.Incident{
				Messages: models.Messages{
					{
						GUID:         "2",
						IncidentGUID: "1",
						Title:        "a title changed",
						Content:      "a content changed",
					},
				},
			}))
			Expect(rr.CheckError()).ToNot(HaveOccurred())

			finalIncident, err := fakeStoreMem.Read("1")
			Expect(err).ToNot(HaveOccurred())
			Expect(finalIncident.GUID).To(Equal("1"))
			Expect(finalIncident.Messages).To(HaveLen(1))
			Expect(finalIncident.Messages[0].Title).To(Equal("a title changed"))
			Expect(finalIncident.Messages[0].Content).To(Equal("a content changed"))
		})
		It("should update only main message when param partial_update_message is set", func() {
			cpns := &models.Components{{
				Name:  Component1.Name,
				Group: Component1.Group,
			}}
			inc1 := models.Incident{
				GUID:       "1",
				CreatedAt:  time.Now().AddDate(0, 0, -2).UTC(),
				UpdatedAt:  time.Now().AddDate(0, 0, -2).UTC(),
				Components: cpns,
				Messages: models.Messages{
					{
						GUID:         "1",
						IncidentGUID: "1",
						Title:        "a title",
						Content:      "a content",
					},
					{
						GUID:         "2",
						IncidentGUID: "1",
						Title:        "a sub title",
						Content:      "a sub content",
					},
				},
				IsScheduled: false,
				State:       models.Monitoring,
			}
			_, err := fakeStoreMem.Create(inc1)
			Expect(err).ToNot(HaveOccurred())

			rr := CallRequest(NewRequestIntAdmin(http.MethodPut, "/v1/incidents/1?partial_update_message=true", models.Incident{
				Messages: models.Messages{
					{
						GUID:    "1",
						Title:   "a title changed",
						Content: "a content changed",
					},
				},
			}))
			Expect(rr.CheckError()).ToNot(HaveOccurred())

			finalIncident, err := fakeStoreMem.Read("1")
			Expect(err).ToNot(HaveOccurred())
			Expect(finalIncident.GUID).To(Equal("1"))
			Expect(finalIncident.Messages).To(HaveLen(2))
			Expect(finalIncident.Messages[0].Title).To(Equal("a title changed"))
			Expect(finalIncident.Messages[0].Content).To(Equal("a content changed"))
		})
	})

	Context("Message api", func() {
		BeforeEach(func() {
			cpns := &models.Components{{
				Name:  Component1.Name,
				Group: Component1.Group,
			}}
			inc1 := models.Incident{
				GUID:       "1",
				CreatedAt:  time.Now().AddDate(0, 0, -2).UTC(),
				UpdatedAt:  time.Now().AddDate(0, 0, -2).UTC(),
				Components: cpns,
				Messages: models.Messages{
					{
						GUID:         "1",
						IncidentGUID: "1",
						Title:        "a title",
						Content:      "a content",
						CreatedAt:    time.Now().AddDate(0, 0, -2).UTC(),
					},
				},
				IsScheduled: false,
				State:       models.Monitoring,
			}
			_, err := fakeStoreMem.Create(inc1)
			Expect(err).ToNot(HaveOccurred())
		})
		Context("AddMessage", func() {
			It("Should give unauthorized when user not set", func() {
				rr := CallRequest(NewRequestInt(http.MethodPost, "/v1/incidents/1/messages", nil))
				Expect(rr.CheckError()).To(HaveOccurred())
				Expect(rr.Code).To(Equal(401))
			})
			It("should add message to incident and emit update", func() {
				rr := CallRequest(NewRequestIntAdmin(http.MethodPost, "/v1/incidents/1/messages", models.Message{
					Title:   "a sub title",
					Content: "a sub content",
				}))
				Expect(rr.CheckError()).ToNot(HaveOccurred())
				Expect(fakeEmitter.EmitCallCount()).To(Equal(1))

				finalIncident, err := fakeStoreMem.Read("1")
				Expect(err).ToNot(HaveOccurred())
				Expect(finalIncident.Messages).To(HaveLen(2))
				lastMessage := finalIncident.LastMessage()
				Expect(lastMessage.GUID).ToNot(BeEmpty())
				Expect(lastMessage.Title).To(Equal("a sub title"))
				Expect(lastMessage.Content).To(Equal("a sub content"))
			})
		})
		Context("DeleteMessage", func() {
			It("Should give unauthorized when user not set", func() {
				rr := CallRequest(NewRequestInt(http.MethodDelete, "/v1/incidents/1/messages/1", nil))
				Expect(rr.CheckError()).To(HaveOccurred())
				Expect(rr.Code).To(Equal(401))
			})
			It("should add message to incident and emit update", func() {
				rr := CallRequest(NewRequestIntAdmin(http.MethodDelete, "/v1/incidents/1/messages/1", nil))
				Expect(rr.CheckError()).ToNot(HaveOccurred())
				Expect(fakeEmitter.EmitCallCount()).To(Equal(1))

				finalIncident, err := fakeStoreMem.Read("1")
				Expect(err).ToNot(HaveOccurred())
				Expect(finalIncident.Messages).To(HaveLen(0))
			})
		})
		Context("ReadMessage", func() {
			It("should add message to incident and emit update", func() {
				rr := CallRequest(NewRequestInt(http.MethodGet, "/v1/incidents/1/messages/1", nil))
				Expect(rr.CheckError()).ToNot(HaveOccurred())

				mess, err := rr.UnmarshalToMessage()
				Expect(err).ToNot(HaveOccurred())

				Expect(mess.GUID).To(Equal("1"))
				Expect(mess.Title).To(Equal("a title"))
				Expect(mess.Content).To(Equal("a content"))

			})
		})
		Context("ReadMessages", func() {
			It("should add message to incident and emit update", func() {
				rr := CallRequest(NewRequestInt(http.MethodGet, "/v1/incidents/1/messages", nil))
				Expect(rr.CheckError()).ToNot(HaveOccurred())

				messages, err := rr.UnmarshalToMessages()
				Expect(err).ToNot(HaveOccurred())

				Expect(messages).To(HaveLen(1))
				mess := messages[0]
				Expect(mess.GUID).To(Equal("1"))
				Expect(mess.Title).To(Equal("a title"))
				Expect(mess.Content).To(Equal("a content"))

			})
		})
		Context("UpdateMessage", func() {
			It("Should give unauthorized when user not set", func() {
				rr := CallRequest(NewRequestInt(http.MethodPut, "/v1/incidents/1/messages/1", nil))
				Expect(rr.CheckError()).To(HaveOccurred())
				Expect(rr.Code).To(Equal(401))
			})
			It("should add message to incident and emit update", func() {
				rr := CallRequest(NewRequestIntAdmin(http.MethodPut, "/v1/incidents/1/messages/1", models.Message{
					Title:   "a title changed",
					Content: "a content changed",
				}))
				Expect(rr.CheckError()).ToNot(HaveOccurred())
				Expect(fakeEmitter.EmitCallCount()).To(Equal(1))

				finalIncident, err := fakeStoreMem.Read("1")
				Expect(err).ToNot(HaveOccurred())
				Expect(finalIncident.Messages).To(HaveLen(1))
				Expect(finalIncident.Messages[0].GUID).To(Equal("1"))
				Expect(finalIncident.Messages[0].Title).To(Equal("a title changed"))
				Expect(finalIncident.Messages[0].Content).To(Equal("a content changed"))
			})
		})
		Context("Notify", func() {
			It("should emit incident for notify with triggerred by user", func() {
				var notifyReq *models.NotifyRequest
				fakeEmitter.EmitStub = func(topic string, args ...interface{}) chan struct{} {
					if topic != "incident" {
						return nil
					}
					notifyReq = args[0].(*models.NotifyRequest)
					return nil
				}

				rr := CallRequest(NewRequestIntAdmin(http.MethodPut, "/v1/incidents/1/notify", nil))

				Expect(notifyReq).ToNot(BeNil())
				Expect(rr.CheckError()).ToNot(HaveOccurred())
				Expect(fakeEmitter.EmitCallCount()).To(Equal(1))
				Expect(notifyReq.TriggerByUser).To(BeTrue())
			})
		})
	})

})
