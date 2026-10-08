package life

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("birth colour", func() {
	It("keeps a colour when all three parents match", func() {
		for _, colour := range []string{"#FF0000", "#00FF00", "#0000FF", "#8C6A21", "#112D4E", "#FFFFFF"} {
			got, err := BirthColour(colour, colour, colour)
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(Equal(colour))
		}
	})

	It("mixes red, green, and blue to a light grey", func() {
		got, err := BirthColour("#FF0000", "#00FF00", "#0000FF")
		Expect(err).NotTo(HaveOccurred())
		Expect(got).NotTo(Equal("#555555"))
		Expect(got).To(Equal("#869290"))
	})
})
