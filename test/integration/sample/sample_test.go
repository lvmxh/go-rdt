package sample_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func FakeServerGet(url string) string {
	urls := map[string]string{
		"/sample": "200"}
	v, ok := urls[url]
	if ok {
		return v
	}
	return "404"
}

// More example: https://github.com/onsi/gomega/tree/master/ghttp
var _ = Describe("Sample", func() {

	var url string
	BeforeEach(func() {
		By("set url")
		url = "/sample"

	})

	JustBeforeEach(func() {
	})

	AfterEach(func() {
	})

	Describe("Gevin a starting http Server.", func() {
		Context("when assess an exist API.", func() {
			It("Should be get 200.", func() {
				Expect(FakeServerGet(url)).To(Equal("200"))
			})
		})

		Context("when assess an non-exist API.", func() {
			It("Should be get 404.", func() {
				Ω(FakeServerGet("error")).Should(Equal("404"))
			})
		})
	})
})
