// This is a example for Third Party Matcher Integrations with ginkgo.
// It is a Mix Matcher with httpexpect and gΩ
// Please ref: http://onsi.github.io/ginkgo/#third-party-integrations
package sample_test

import (
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/gavv/httpexpect.v1"
	"gopkg.in/gavv/httpexpect.v1/_examples"

	"openstackcore-rdtagent/test/integration/test_helpers"
)

// more example: https://github.com/gavv/httpexpect/tree/master/_examples
var _ = Describe("Sample", func() {

	// For Global Setup and Teardown ref:
	// http://onsi.github.io/ginkgo/#global-setup-and-teardown-beforesuite-and-aftersuite
	var (
		v1url   string
		hexpect *httpexpect.Expect
		handler http.Handler
		server  *httptest.Server
	)

	BeforeEach(func() {
		By("set url")
		v1url = testhelpers.GetV1URL()
	})

	JustBeforeEach(func() {
		// create http.Handler
		handler = examples.FruitsHandler()

		// run server using httptest
		server = httptest.NewServer(handler)

		// create httpexpect instance
		hexpect = httpexpect.New(GinkgoT(), server.URL)
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("Gevin a fresh fruits Web Server.", func() {
		Context("when request '/fruits'  API.", func() {
			It("Should be get 200 and should not get any fruits.", func() {
				hexpect.GET("/fruits").
					Expect().
					Status(http.StatusOK).JSON().Array().Empty()
			})
		})
	})

	Describe("Gevin a really RMD runtime evn.", func() {
		Context("Configure file is generated with hard code", func() {
			It("The server should set v1url correctly", func() {
				Expect(v1url).To(Equal("http://localhost:8088/v1/"))
			})
		})
	})
})
