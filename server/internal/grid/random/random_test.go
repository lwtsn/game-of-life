package random

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("random source", func() {
	It("fills the whole grid", func() {
		first := newRandomSeed(1).Next(nil)
		second := newRandomSeed(2).Next(nil)

		Expect(first.Width()).To(Equal(cols))
		Expect(first.Height()).To(Equal(rows))
		Expect(first.Cells()).To(HaveLen(cols * rows))
		Expect(first.Cells()).To(ContainElement(1))
		Expect(first.Cells()).To(ContainElement(0))
		for _, cell := range first.Cells() {
			Expect(cell).To(BeElementOf(0, 1))
		}
		Expect(first.Cells()).NotTo(Equal(second.Cells()))
	})
})
