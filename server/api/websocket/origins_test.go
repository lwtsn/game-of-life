package websocket

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("origins", func() {
	DescribeTable("defaults to this machine",
		func(origin string, want bool) {
			Expect(ParseOrigins("").Allows(origin)).To(Equal(want))
		},
		Entry("vite dev", "http://localhost:5173", true),
		Entry("loopback", "http://127.0.0.1:8080", true),
		Entry("no port", "http://localhost", true),
		Entry("another site", "https://example.com", false),
		Entry("not http", "file://localhost", false),
		Entry("garbage", "::", false),
	)

	DescribeTable("uses the configured list",
		func(origin string, want bool) {
			origins := ParseOrigins(" game.example.com , *.example.org:* ")
			Expect(origins.Allows(origin)).To(Equal(want))
		},
		Entry("exact host", "https://game.example.com", true),
		Entry("exact host, other port", "https://game.example.com:8443", false),
		Entry("wildcard host and port", "http://play.example.org:5173", true),
		Entry("case is ignored", "https://GAME.example.com", true),
		Entry("localhost is no longer allowed", "http://localhost:5173", false),
	)

	It("treats an unset list as the defaults", func() {
		var origins Origins
		Expect(origins.Patterns()).To(Equal([]string(DefaultOrigins)))
	})
})
