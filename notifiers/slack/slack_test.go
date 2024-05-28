package slack_test

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	"github.com/orange-cloudfoundry/statusetat/config"
	"github.com/orange-cloudfoundry/statusetat/models"
	"github.com/orange-cloudfoundry/statusetat/notifiers"

	"github.com/orange-cloudfoundry/statusetat/notifiers/slack"
)

var _ = Describe("Slack", func() {
	var notifier notifiers.Notifier
	var server *ghttp.Server

	BeforeEach(func() {
		var err error
		server = ghttp.NewServer()
		notifier, err = (&slack.Slack{}).Creator(map[string]interface{}{
			"endpoint": server.URL(),
			"channel":  "my channel",
		}, config.BaseInfo{})
		if err != nil {
			panic(err)
		}

	})
	AfterEach(func() {
		// shut down the server between tests
		server.Close()
	})
	Context("Creator", func() {
		It("should create notifier", func() {
			newNotif, err := (&slack.Slack{}).Creator(map[string]interface{}{
				"endpoint": "http://slack.com/incoming",
				"channel":  "my channel",
			}, config.BaseInfo{})
			Expect(err).To(BeNil())
			Expect(newNotif).NotTo(BeNil())
		})
	})
	Context("Notify", func() {
		Context("Is scheduled task", func() {
			It("should wrote on slack a scheduled task message", func() {
				var slackReq slack.SlackRequest
				server.RouteToHandler("POST", "/", func(writer http.ResponseWriter, request *http.Request) {
					b, _ := io.ReadAll(request.Body)
					err := json.Unmarshal(b, &slackReq)
					Expect(err).To(BeNil())
				})

				notifyReq := models.NewNotifyRequest(models.Incident{
					GUID: "aguid",
					Messages: []models.Message{
						{
							CreatedAt: time.Time{},
							Title:     "A title",
							Content:   "content",
						},
					},
					IsScheduled: true,
				}, false)

				err := notifier.Notify(notifyReq)

				Expect(err).ToNot(HaveOccurred())
				Expect(slackReq.Channel).To(Equal("my channel"))
				Expect(slackReq.Username).To(ContainSubstring("Scheduled maintenance"))
				Expect(slackReq.Attachments).To(HaveLen(1))
				Expect(slackReq.Attachments[0].Title).To(Equal("A title"))
				Expect(slackReq.Attachments[0].Text).To(Equal("content"))
				Expect(slackReq.Attachments[0].Fields).To(HaveLen(3))

			})
		})
		Context("Is incident", func() {
			It("should wrote on slack a scheduled task message", func() {
				var slackReq slack.SlackRequest
				server.RouteToHandler("POST", "/", func(writer http.ResponseWriter, request *http.Request) {
					b, _ := io.ReadAll(request.Body)
					err := json.Unmarshal(b, &slackReq)
					Expect(err).To(BeNil())
				})

				notifyReq := models.NewNotifyRequest(models.Incident{
					GUID: "aguid",
					Messages: []models.Message{
						{
							CreatedAt: time.Time{},
							Title:     "A title",
							Content:   "content",
						},
					},
					IsScheduled: false,
				}, false)

				err := notifier.Notify(notifyReq)

				Expect(err).ToNot(HaveOccurred())
				Expect(slackReq.Channel).To(Equal("my channel"))
				Expect(slackReq.Username).To(ContainSubstring("Incident"))
				Expect(slackReq.Attachments).To(HaveLen(1))
				Expect(slackReq.Attachments[0].Title).To(ContainSubstring("A title"))
				Expect(slackReq.Attachments[0].Text).To(Equal("content"))
				Expect(slackReq.Attachments[0].Fields).To(HaveLen(3))

			})
			Context("multiple messages", func() {
				Context("not triggerred by user", func() {
					It("should not wrote on slack update", func() {
						var slackReq slack.SlackRequest
						server.RouteToHandler("POST", "/", func(writer http.ResponseWriter, request *http.Request) {
							b, _ := io.ReadAll(request.Body)
							err := json.Unmarshal(b, &slackReq)
							Expect(err).To(BeNil())
						})

						notifyReq := models.NewNotifyRequest(models.Incident{
							GUID: "aguid",
							Messages: []models.Message{
								{
									CreatedAt: time.Time{},
									Title:     "A title",
									Content:   "content",
								},
								{
									CreatedAt: time.Time{},
									Title:     "A second title",
									Content:   "another content",
								},
							},
							IsScheduled: false,
						}, false)

						err := notifier.Notify(notifyReq)

						Expect(err).ToNot(HaveOccurred())
						Expect(slackReq.Channel).To(Equal(""))
						Expect(slackReq.Username).To(ContainSubstring(""))
						Expect(slackReq.Attachments).To(HaveLen(0))

					})
				})
				Context("triggerred by user", func() {
					It("should wrote on slack update", func() {
						var slackReq slack.SlackRequest
						server.RouteToHandler("POST", "/", func(writer http.ResponseWriter, request *http.Request) {
							b, _ := io.ReadAll(request.Body)
							err := json.Unmarshal(b, &slackReq)
							Expect(err).To(BeNil())
						})

						notifyReq := models.NewNotifyRequest(models.Incident{
							GUID: "aguid",
							Messages: []models.Message{
								{
									CreatedAt: time.Time{},
									Title:     "A second title",
									Content:   "another content",
								},
								{
									CreatedAt: time.Time{},
									Title:     "A title",
									Content:   "content",
								},
							},
							IsScheduled: false,
						}, true)

						err := notifier.Notify(notifyReq)

						Expect(err).ToNot(HaveOccurred())
						Expect(slackReq.Channel).To(Equal("my channel"))
						Expect(slackReq.Username).To(ContainSubstring("Incident"))
						Expect(slackReq.Attachments).To(HaveLen(1))
						Expect(slackReq.Attachments[0].Title).To(ContainSubstring("A second title"))
						Expect(slackReq.Attachments[0].Text).To(Equal("another content"))
						Expect(slackReq.Attachments[0].Fields).To(HaveLen(3))

					})
				})
			})
		})
	})
})
