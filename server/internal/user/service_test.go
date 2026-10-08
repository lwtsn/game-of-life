package user

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("service", func() {
	It("returns the colour for an IP", func() {
		svc := NewService()
		person := svc.ByIP("::ffff:198.51.100.10")

		Expect(person.IP()).To(Equal("198.51.100.10"))
		Expect(person.Colour()).To(Equal(New("198.51.100.10").Colour()))
		Expect(svc.Colour("198.51.100.10")).To(Equal(person.Colour()))
		Expect(svc.Colour("198.51.100.11")).NotTo(Equal(person.Colour()))
	})

	It("keeps one user per address", func() {
		svc := NewService()
		first := svc.Join("::ffff:198.51.100.10")
		again := svc.Join("198.51.100.10")

		Expect(again.IP()).To(Equal("198.51.100.10"))
		Expect(again.Colour()).To(Equal(first.Colour()))
		Expect(svc.Colours()).To(Equal([]string{first.Colour()}))

		svc.Join("198.51.100.11")
		Expect(svc.Colours()).To(HaveLen(2))

		svc.Leave("::ffff:198.51.100.10")
		Expect(svc.Colours()).To(Equal([]string{New("198.51.100.11").Colour()}))
	})
})
