package grid

import (
	"context"
	"time"

	"game_of_life/server/internal/grid/source"
	"game_of_life/server/internal/grid/source/mocks"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
)

type stubFrame struct {
	width  int
	height int
	cells  []int
}

func (f stubFrame) Width() int   { return f.width }
func (f stubFrame) Height() int  { return f.height }
func (f stubFrame) Cells() []int { return f.cells }
func (f stubFrame) ToJson() ([]byte, error) {
	return source.Encode(f)
}

var _ = Describe("Grid", func() {
	It("stores a copy of the current cells", func() {
		cells := []int{1, 0, 1, 0}
		frame := stubFrame{width: 2, height: 2, cells: cells}
		src := mocks.NewMockSource(GinkgoT())
		src.EXPECT().Next(nil).Return(frame).Once()

		board := newGrid(src)
		cells[0] = 9

		got := board.Current()
		Expect(got.Width()).To(Equal(2))
		Expect(got.Height()).To(Equal(2))
		Expect(got.Cells()).To(Equal([]int{1, 0, 1, 0}))
	})

	It("passes the stored board to the source", func() {
		first := stubFrame{width: 2, height: 2, cells: []int{1, 0, 0, 0}}
		second := stubFrame{width: 2, height: 2, cells: []int{0, 1, 0, 0}}
		src := mocks.NewMockSource(GinkgoT())
		src.EXPECT().Next(nil).Return(first).Once()
		src.EXPECT().Next(mock.MatchedBy(func(current source.Frame) bool {
			cells := current.Cells()
			return len(cells) == 4 && cells[0] == 1 && cells[1] == 0
		})).Return(second).Once()

		board := newGrid(src)
		board.Advance()

		Expect(board.Current().Cells()).To(Equal([]int{0, 1, 0, 0}))
	})

	It("publishes the current board when started", func() {
		frame := stubFrame{width: 2, height: 2, cells: []int{1, 0, 0, 1}}
		src := mocks.NewMockSource(GinkgoT())
		src.EXPECT().Next(nil).Return(frame).Once()

		board := newGrid(src)
		ctx, cancel := context.WithCancel(context.Background())
		DeferCleanup(cancel)
		board.Start(ctx)

		want, err := frame.ToJson()
		Expect(err).NotTo(HaveOccurred())
		Eventually(board.Updates(), time.Second, 10*time.Millisecond).Should(Receive(Equal(want)))
	})
})
