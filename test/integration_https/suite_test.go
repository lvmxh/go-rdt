// +build integration_https
package integration_https

import (
	"crypto/tls"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/gavv/httpexpect.v1"
	"net/http"
	"openstackcore-rdtagent/test/test_helpers"
	"os"
	"testing"
)

var (
	he *httpexpect.Expect
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration_HTTPS")
}

var _ = BeforeSuite(func() {

	err := testhelpers.ConfigInit(os.Getenv("CONF"))
	Expect(err).NotTo(HaveOccurred())

	skipVerify := false
	if testhelpers.GetClientAuthType() == "no" {
		skipVerify = true
	}

	he = httpexpect.WithConfig(httpexpect.Config{
		BaseURL: testhelpers.GetHTTPSV1URL(),
		Client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: skipVerify},
			},
		},
		Reporter: httpexpect.NewAssertReporter(
			httpexpect.NewAssertReporter(GinkgoT()),
		),
		Printers: []httpexpect.Printer{
			httpexpect.NewDebugPrinter(GinkgoT(), true),
		},
	})

})
