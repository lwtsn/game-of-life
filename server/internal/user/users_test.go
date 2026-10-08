package user

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("users", func() {
	It("keeps one user per address", func() {
		people := NewUsers()
		first := people.Join("::ffff:198.51.100.10")
		again := people.Join("198.51.100.10")

		Expect(again.IP()).To(Equal("198.51.100.10"))
		Expect(again.IP()).To(Equal(first.IP()))
		Expect(again.Colour()).To(Equal(first.Colour()))
		Expect(people.Colours()).To(Equal([]string{first.Colour()}))

		people.Join("198.51.100.11")
		Expect(people.Colours()).To(HaveLen(2))

		people.Leave("::ffff:198.51.100.10")
		Expect(people.Colours()).To(Equal([]string{New("198.51.100.11").Colour()}))
	})
})
