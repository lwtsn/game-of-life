package grid

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestRandomFillsTheWholeGrid(t *testing.T) {
	g := NewWithT(t)

	first := NewRandomSeed(1).Next()
	second := NewRandomSeed(2).Next()

	g.Expect(first.Width).To(Equal(Cols))
	g.Expect(first.Height).To(Equal(Rows))
	g.Expect(first.Cells).To(HaveLen(Cols * Rows))
	g.Expect(first.Cells).To(ContainElement(1))
	g.Expect(first.Cells).To(ContainElement(0))
	for _, cell := range first.Cells {
		g.Expect(cell).To(BeElementOf(0, 1))
	}
	g.Expect(first.Cells).NotTo(Equal(second.Cells))
}
