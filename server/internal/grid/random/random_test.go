package random

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestRandomFillsTheWholeGrid(t *testing.T) {
	g := NewWithT(t)

	first := newRandomSeed(1).Next(nil)
	second := newRandomSeed(2).Next(nil)

	g.Expect(first.Width()).To(Equal(cols))
	g.Expect(first.Height()).To(Equal(rows))
	g.Expect(first.Cells()).To(HaveLen(cols * rows))
	g.Expect(first.Cells()).To(ContainElement(1))
	g.Expect(first.Cells()).To(ContainElement(0))
	for _, cell := range first.Cells() {
		g.Expect(cell).To(BeElementOf(0, 1))
	}
	g.Expect(first.Cells()).NotTo(Equal(second.Cells()))
}
