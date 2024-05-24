package storages_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestStorages(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Storages Suite")
}
