package serves_test

import (
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Subscribe", func() {
	Context("SubscribeEmail", func() {
		It("should store new subscriber", func() {
			email := "toto@toto.com"
			rr := CallRequest(NewRequestInt(http.MethodPost, "/v1/subscribe?email="+email, nil))
			Expect(rr.CheckError()).ToNot(HaveOccurred())
			Expect(fakeStoreMem.SubscribeCallCount()).To(Equal(1))
			subs, err := fakeStoreMem.Subscribers()
			Expect(err).ToNot(HaveOccurred())
			Expect(subs).To(HaveLen(1))
			Expect(subs[0]).To(Equal(email))
		})
	})

	Context("UnsubscribeEmail", func() {
		It("should remove subscriber", func() {
			email := "toto@toto.com"
			err := fakeStoreMem.Subscribe(email)
			Expect(err).ToNot(HaveOccurred())
			subs, err := fakeStoreMem.Subscribers()
			Expect(err).ToNot(HaveOccurred())
			Expect(subs).To(HaveLen(1))
			Expect(subs[0]).To(Equal(email))

			rr := CallRequest(NewRequestInt(http.MethodPost, "/v1/unsubscribe?email="+email, nil))
			Expect(rr.CheckError()).ToNot(HaveOccurred())
			Expect(fakeStoreMem.UnsubscribeCallCount()).To(Equal(1))
			subs, err = fakeStoreMem.Subscribers()
			Expect(err).ToNot(HaveOccurred())
			Expect(subs).To(HaveLen(0))
		})
	})
})
