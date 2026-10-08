package service

import (
	"game_of_life/server/internal/user"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("user service", func() {
	It("returns the colour for an IP", func() {
		svc := New()
		person := svc.ByIP("::ffff:198.51.100.10")

		Expect(person.IP()).To(Equal("198.51.100.10"))
		Expect(person.Colour()).To(Equal(user.New("198.51.100.10").Colour()))
		Expect(svc.Colour("198.51.100.10")).To(Equal(person.Colour()))
		Expect(svc.Colour("198.51.100.11")).NotTo(Equal(person.Colour()))
	})

	It("joins and leaves the current set", func() {
		svc := New()
		person := svc.Join("::ffff:198.51.100.10")

		Expect(person.IP()).To(Equal("198.51.100.10"))
		Expect(svc.Colours()).To(Equal([]string{person.Colour()}))

		svc.Leave("198.51.100.10")
		Expect(svc.Colours()).To(BeEmpty())
	})
})
