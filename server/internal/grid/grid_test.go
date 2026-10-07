package grid

import (
	"testing"

	"game_of_life/server/internal/grid/source/mocks"
	. "github.com/onsi/gomega"
)

type stubFrame struct {
	width  int
	height int
	cells  []int
}

func (f stubFrame) Width() int   { return f.width }
func (f stubFrame) Height() int  { return f.height }
func (f stubFrame) Cells() []int { return f.cells }

func TestGridReturnsTheSourceValues(t *testing.T) {
	g := NewWithT(t)

	cells := []int{1, 0, 1, 0}
	frame := stubFrame{width: 2, height: 2, cells: cells}
	src := mocks.NewMockSource(t)
	src.EXPECT().Next().Return(frame).Once()

	got := newGrid(src).Next()

	g.Expect(got.Width()).To(Equal(2))
	g.Expect(got.Height()).To(Equal(2))
	g.Expect(got.Cells()).To(Equal(cells))
}
