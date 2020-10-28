package slack_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSlack(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Slack Suite")
}
