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
		Describe("Get policy", func() {

			BeforeEach(func() {
				path = "/policy"
			})

			Context("Get policy with valid Berkely db credentials", func() {
				BeforeEach(func() {
					username = "user"
					password = "user1"
				})
				It("Should return 200OK", func() {
					he.GET(path).
						WithHeader("Content-Type", "application/json").
						WithBasicAuth(username, password).
						Expect().
						Status(http.StatusOK)
				})
			})

			Context("Get policy with invalid Berkely db user", func() {
				BeforeEach(func() {
					username = "use"
					password = "user1"
				})
				It("Should return 401 StatusUnauthorized", func() {
					he.GET(path).
						WithHeader("Content-Type", "application/json").
						WithBasicAuth(username, password).
						Expect().
						Status(http.StatusUnauthorized).
						Text().
						Equal("User not known to the underlying authentication module\n")
				})
			})

			Context("Get policy with incorrect Berkely db credentials", func() {
				BeforeEach(func() {
					username = "user"
					password = "user2"
				})
				It("Should return 401 StatusUnauthorized", func() {
					he.GET(path).
						WithHeader("Content-Type", "application/json").
						WithBasicAuth(username, password).
						Expect().
						Status(http.StatusUnauthorized).
						Text().
						Equal("Authentication failure\n")
				})
			})

			Context("Get policy with valid unix credentials", func() {
				// Please use credentials different from those defined in Berkely db for a consistent error message
				BeforeEach(func() {
					username = "root"
					password = "s"
				})
				It("Should return 200OK", func() {
					he.GET(path).
						WithHeader("Content-Type", "application/json").
						WithBasicAuth(username, password).
						Expect().
						Status(http.StatusOK)
				})
			})

			Context("Get policy with invalid unix user", func() {
				// Please use credentials different from those defined in Berkely db for a consistent error message
				BeforeEach(func() {
					username = "com"
					password = "common"
				})
				It("Should return 401 StatusUnauthorized", func() {
					he.GET(path).
						WithHeader("Content-Type", "application/json").
						WithBasicAuth(username, password).
						Expect().
						Status(http.StatusUnauthorized).
						Text().
						Equal("User not known to the underlying authentication module\n")
				})
			})

			Context("Get policy with incorrect unix credentials", func() {
				// Please use credentials different from those defined in Berkely db for a consistent error message
				BeforeEach(func() {
					username = "root"
					password = "com"
				})
				It("Should return 401 StatusUnauthorized", func() {
					he.GET(path).
						WithHeader("Content-Type", "application/json").
						WithBasicAuth(username, password).
						Expect().
						Status(http.StatusUnauthorized).
						Text().
						Equal("User not known to the underlying authentication module\n")
				})
			})
		})
	})
})
