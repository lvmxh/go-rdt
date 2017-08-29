package sample_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"openstackcore-rdtagent/test/integration/test_helpers"
)

func TestSrc(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Sample Suite")
}

var _ = BeforeSuite(func() {
	err := testhelpers.ConfigInit(os.Getenv("CONF"))
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
})
