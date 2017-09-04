package integration_test

import (
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo"

	"gopkg.in/gavv/httpexpect.v1"
	"openstackcore-rdtagent/test/integration/test_helpers"
)

var cacheSchemaTemplate string = `{
	"type": "object",
	"properties": {
		"rdt": %s,
		"cqm": %s,
		"cdp": %s,
		"cdp_enable": %s,
		"cat": %s,
		"cat_enable": %s,
		"caches": {
			"type": "object",
			"properties": {
				"l3": {
					"type": "object",
					"properties": {
						"number": %s,
						"cache_ids": {"type": "array", "items": %s}
					}
				},
				"l2": {
					"type": "object",
					"properties": {
						"number": %s,
						"cache_ids": {"type": "array", "items": %s}
					}
				}
			}
		}
	},
	"required": ["rdt", "cqm", "cdp", "cdp_enable", "cat", "cat_enable", "caches"]
}`

var cacheSchema string = fmt.Sprintf(cacheSchemaTemplate,
	testhelpers.BoolSchema,
	testhelpers.BoolSchema,
	testhelpers.BoolSchema,
	testhelpers.BoolSchema,
	testhelpers.BoolSchema,
	testhelpers.BoolSchema,
	testhelpers.PositiveInteger,
	testhelpers.NonNegativeInteger,
	testhelpers.PositiveInteger,
	testhelpers.NonNegativeInteger)

var _ = Describe("Caches", func() {

	var (
		v1url string
		he    *httpexpect.Expect
	)

	BeforeEach(func() {
		By("set url")
		v1url = testhelpers.GetV1URL()
		he = httpexpect.New(GinkgoT(), v1url)
	})

	AfterEach(func() {
	})

	Describe("Get the new system", func() {
		Context("when request 'cache' API", func() {
			It("Should be return 200", func() {

				repos := he.GET("/cache").
					WithHeader("Content-Type", "application/json").
					Expect().
					Status(http.StatusOK).
					JSON()

				repos.Schema(cacheSchema)
			})
		})
	})
})
