// +build integration_https
package integration_https

import (
	. "github.com/onsi/ginkgo"
	"net/http"
)

var _ = Describe("PAMAuth", func() {

	var (
		path, username, password string
	)

	Describe("Get https requests", func() {
		BeforeEach(func() {
			username = "user"
			password = "user1"
		})

		Context("Get policy", func() {
			BeforeEach(func() {
				path = "/policy"
			})
			It("Should return 200OK", func() {
				he.GET(path).
					WithHeader("Content-Type", "application/json").
					WithBasicAuth(username, password).
					Expect().
					Status(http.StatusOK)
			})
		})
	})
})
