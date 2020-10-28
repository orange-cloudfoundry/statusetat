package slack_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/orange-cloudfoundry/statusetat/config"
	"github.com/orange-cloudfoundry/statusetat/models"
	"github.com/orange-cloudfoundry/statusetat/notifiers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

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
					b, _ := ioutil.ReadAll(request.Body)
					err := json.Unmarshal(b, &slackReq)
					Expect(err).To(BeNil())
				})
				notifier.Notify(models.Incident{
					GUID: "aguid",
					Messages: []models.Message{
						{
							CreatedAt: time.Time{},
							Title:     "A title",
							Content:   "content",
						},
					},
					IsScheduled: true,
				})

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
					b, _ := ioutil.ReadAll(request.Body)
					err := json.Unmarshal(b, &slackReq)
					Expect(err).To(BeNil())
				})
				notifier.Notify(models.Incident{
					GUID: "aguid",
					Messages: []models.Message{
						{
							CreatedAt: time.Time{},
							Title:     "A title",
							Content:   "content",
						},
					},
					IsScheduled: false,
				})

				Expect(slackReq.Channel).To(Equal("my channel"))
				Expect(slackReq.Username).To(ContainSubstring("Incident"))
				Expect(slackReq.Attachments).To(HaveLen(1))
				Expect(slackReq.Attachments[0].Title).To(ContainSubstring("A title"))
				Expect(slackReq.Attachments[0].Text).To(Equal("content"))
				Expect(slackReq.Attachments[0].Fields).To(HaveLen(3))

			})
		})
	})
})
