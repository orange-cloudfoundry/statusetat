package email_test

import (
	"bytes"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/gomail.v2"

	"github.com/orange-cloudfoundry/statusetat/config"
	"github.com/orange-cloudfoundry/statusetat/models"
	"github.com/orange-cloudfoundry/statusetat/notifiers"
	"github.com/orange-cloudfoundry/statusetat/notifiers/email"
	"github.com/orange-cloudfoundry/statusetat/notifiers/email/emailfakes"
)

var _ = Describe("Email", func() {
	var notifier notifiers.Notifier
	var fakeDialer *emailfakes.FakeEmailDialer

	BeforeEach(func() {
		var err error
		fakeDialer = &emailfakes.FakeEmailDialer{}
		notifier, err = (&email.Email{}).Creator(map[string]interface{}{
			"host": "toto.com",
			"port": 993,
			"from": "me@me.com",
		}, config.BaseInfo{})
		if err != nil {
			panic(err)
		}
		notifier.(*email.Email).SetDialer(fakeDialer)
	})
	AfterEach(func() {
	})
	Context("Creator", func() {
		It("should create notifier", func() {
			newNotif, err := (&email.Email{}).Creator(map[string]interface{}{
				"host": "toto.com",
				"port": 993,
				"from": "me@me.com",
			}, config.BaseInfo{})
			Expect(err).To(BeNil())
			Expect(newNotif).NotTo(BeNil())
		})
	})
	Context("Notify", func() {
		Context("Is scheduled task", func() {
			It("should wrote on slack a scheduled task message if state is idle", func() {
				subscriber := "user@user.com"

				fakeDialer.DialAndSendStub = func(message ...*gomail.Message) error {
					Expect(message).To(HaveLen(1))
					m := message[0]
					Expect(m.GetHeader("From")).To(ContainElement("me@me.com"))
					Expect(m.GetHeader("To")).To(ContainElement(subscriber))
					Expect(m.GetHeader("Subject")).To(HaveLen(1))
					Expect(m.GetHeader("Subject")[0]).To(ContainSubstring("Scheduled task]"))
					Expect(m.GetHeader("Auto-submitted")).To(ContainElement("auto-generated"))
					buf := &bytes.Buffer{}
					_, err := m.WriteTo(buf)
					Expect(err).ToNot(HaveOccurred())

					Expect(buf.String()).To(ContainSubstring("<h1>Scheduled Maintenance"))
					Expect(buf.String()).To(ContainSubstring("Planned Duration"))
					return nil
				}

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
					State:       models.Idle,
				}, false)
				notifyReq.Subscribers = []string{subscriber}

				err := notifier.Notify(notifyReq)
				Expect(err).ToNot(HaveOccurred())
				Expect(fakeDialer.DialAndSendCallCount()).To(Equal(1))
			})

			It("should wrote on slack a scheduled task started message if state is unresolved", func() {
				subscriber := "user@user.com"

				fakeDialer.DialAndSendStub = func(message ...*gomail.Message) error {
					Expect(message).To(HaveLen(1))
					m := message[0]
					Expect(m.GetHeader("From")).To(ContainElement("me@me.com"))
					Expect(m.GetHeader("To")).To(ContainElement(subscriber))
					Expect(m.GetHeader("Subject")).To(HaveLen(1))
					Expect(m.GetHeader("Subject")[0]).To(ContainSubstring("Scheduled task has started]"))
					Expect(m.GetHeader("Auto-submitted")).To(ContainElement("auto-generated"))
					buf := &bytes.Buffer{}
					_, err := m.WriteTo(buf)
					Expect(err).ToNot(HaveOccurred())

					Expect(buf.String()).To(ContainSubstring("<h1>Scheduled Maintenance"))
					Expect(buf.String()).To(ContainSubstring("Planned Duration"))
					return nil
				}

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
					State:       models.Unresolved,
				}, false)
				notifyReq.Subscribers = []string{subscriber}

				err := notifier.Notify(notifyReq)
				Expect(err).ToNot(HaveOccurred())
				Expect(fakeDialer.DialAndSendCallCount()).To(Equal(1))
			})

			It("should wrote on slack a scheduled task finished message if state is resolved", func() {
				subscriber := "user@user.com"

				fakeDialer.DialAndSendStub = func(message ...*gomail.Message) error {
					Expect(message).To(HaveLen(1))
					m := message[0]
					Expect(m.GetHeader("From")).To(ContainElement("me@me.com"))
					Expect(m.GetHeader("To")).To(ContainElement(subscriber))
					Expect(m.GetHeader("Subject")).To(HaveLen(1))
					Expect(m.GetHeader("Subject")[0]).To(ContainSubstring("Scheduled task has finished]"))
					Expect(m.GetHeader("Auto-submitted")).To(ContainElement("auto-generated"))
					buf := &bytes.Buffer{}
					_, err := m.WriteTo(buf)
					Expect(err).ToNot(HaveOccurred())

					Expect(buf.String()).To(ContainSubstring("<h1>Scheduled Maintenance"))
					Expect(buf.String()).To(ContainSubstring("Final Duration"))
					return nil
				}

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
					State:       models.Resolved,
				}, false)
				notifyReq.Subscribers = []string{subscriber}

				err := notifier.Notify(notifyReq)
				Expect(err).ToNot(HaveOccurred())
			})

		})
		Context("Is incident", func() {
			It("should wrote on slack a scheduled task message", func() {
				subscriber := "user@user.com"

				fakeDialer.DialAndSendStub = func(message ...*gomail.Message) error {
					Expect(message).To(HaveLen(1))
					m := message[0]
					Expect(m.GetHeader("From")).To(ContainElement("me@me.com"))
					Expect(m.GetHeader("To")).To(ContainElement(subscriber))
					Expect(m.GetHeader("Subject")).To(HaveLen(1))
					Expect(m.GetHeader("Subject")[0]).To(ContainSubstring("Incident]"))
					Expect(m.GetHeader("Auto-submitted")).To(ContainElement("auto-generated"))
					buf := &bytes.Buffer{}
					_, err := m.WriteTo(buf)
					Expect(err).ToNot(HaveOccurred())

					Expect(buf.String()).ToNot(ContainSubstring("<h1>Scheduled Task"))
					return nil
				}

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
				notifyReq.Subscribers = []string{subscriber}

				err := notifier.Notify(notifyReq)
				Expect(err).ToNot(HaveOccurred())
				Expect(fakeDialer.DialAndSendCallCount()).To(Equal(1))
			})
		})
		Context("have many subscribers", func() {
			It("should send to all subscribers", func() {

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
				notifyReq.Subscribers = []string{"user@user.com", "user2@user.com", "user3@user.com"}

				err := notifier.Notify(notifyReq)
				Expect(err).ToNot(HaveOccurred())
				Expect(fakeDialer.DialAndSendCallCount()).To(Equal(3))
			})
		})
	})
})
