// This is a example for Third Party Matcher Integrations with ginkgo.
// Please ref: http://onsi.github.io/ginkgo/#third-party-integrations
package sample_test

import (
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo"
	"gopkg.in/gavv/httpexpect.v1"
	"gopkg.in/gavv/httpexpect.v1/_examples"
)

// more example: https://github.com/gavv/httpexpect/tree/master/_examples
var _ = Describe("Sample", func() {

	// For Global Setup and Teardown ref:
	// http://onsi.github.io/ginkgo/#global-setup-and-teardown-beforesuite-and-aftersuite
	var (
		url     string
		hexpect *httpexpect.Expect
		handler http.Handler
		server  *httptest.Server
	)

	BeforeEach(func() {
		By("set url")
		url = "/fruits"

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
})
