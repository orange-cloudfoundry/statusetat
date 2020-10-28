package email_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestEmail(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Email Suite")
}
