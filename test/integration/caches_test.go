package integration_test

import (
	"net/http"

	. "github.com/onsi/ginkgo"

	"gopkg.in/gavv/httpexpect.v1"
	"openstackcore-rdtagent/test/integration/test_helpers"
)

var cacheSchemaTemplate string = `{
	"type": "object",
	"properties": {
		"rdt": {{.bool}},
		"cqm": {{.bool}},
		"cdp": {{.bool}},
		"cdp_enable": {{.bool}},
		"cat": {{.bool}},
		"cat_enable": {{.bool}},
		"caches": {
			"type": "object",
			"properties": {
				"l3": {
					"type": "object",
					"properties": {
						"number": {{.pint}},
						"cache_ids": {"type": "array", "items": {{.uint}}}
					}
				},
				"l2": {
					"type": "object",
					"properties": {
						"number": {{.pint}},
						"cache_ids": {"type": "array", "items": {{.uint}}}
					}
				}
			}
		}
	},
	"required": ["rdt", "cqm", "cdp", "cdp_enable", "cat", "cat_enable", "caches"]
}`

var _ = Describe("Caches", func() {

	var (
		v1url       string
		he          *httpexpect.Expect
		cacheSchema string
	)

	BeforeEach(func() {
		By("set url")
		v1url = testhelpers.GetV1URL()
		he = httpexpect.New(GinkgoT(), v1url)
		cacheSchema, _ = testhelpers.FormatByKey(cacheSchemaTemplate,
			map[string]interface{}{
				"bool": testhelpers.BoolSchema,
				"pint": testhelpers.PositiveInteger,
				"uint": testhelpers.NonNegativeInteger})
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
