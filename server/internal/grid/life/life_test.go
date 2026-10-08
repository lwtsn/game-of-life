package life

import (
	"game_of_life/server/internal/grid/source"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
)

var _ = Describe("life", func() {
	var src source.Source

	BeforeEach(func() {
		app := fx.New(Module, fx.Populate(&src), fx.NopLogger)
		Expect(app.Err()).NotTo(HaveOccurred())
	})

	DescribeTable("the next generation",
		func(width, height int, cells, want []int) {
			current := snapshot{width: width, height: height, cells: cells}
			got := src.Next(current)

			Expect(current.Cells()).To(Equal(cells))
			Expect(got.Width()).To(Equal(width))
			Expect(got.Height()).To(Equal(height))
			Expect(got.Cells()).To(Equal(want))
		},
		Entry("a dead board stays dead",
			2, 2,
			[]int{0, 0, 0, 0},
			[]int{0, 0, 0, 0},
		),
		Entry("a live cell with no neighbours dies",
			1, 1,
			[]int{1},
			[]int{0},
		),
		Entry("two adjacent live cells die",
			2, 1,
			[]int{1, 1},
			[]int{0, 0},
		),
		Entry("a live cell with two neighbours lives",
			3, 3,
			[]int{
				1, 0, 0,
				0, 1, 0,
				0, 0, 1,
			},
			[]int{
				0, 0, 0,
				0, 1, 0,
				0, 0, 0,
			},
		),
		Entry("a live cell with three neighbours lives",
			2, 2,
			[]int{
				1, 1,
				1, 1,
			},
			[]int{
				1, 1,
				1, 1,
			},
		),
		Entry("a live cell with four neighbours dies",
			3, 3,
			[]int{
				1, 1, 1,
				1, 1, 0,
				0, 0, 0,
			},
			[]int{
				1, 0, 1,
				1, 0, 1,
				0, 0, 0,
			},
		),
		Entry("a dead cell with three neighbours becomes live",
			3, 3,
			[]int{
				1, 1, 0,
				1, 0, 0,
				0, 0, 0,
			},
			[]int{
				1, 1, 0,
				1, 1, 0,
				0, 0, 0,
			},
		),
		Entry("a dead cell with two neighbours stays dead",
			2, 2,
			[]int{
				1, 1,
				0, 0,
			},
			[]int{
				0, 0,
				0, 0,
			},
		),
		Entry("a dead cell with four neighbours stays dead",
			3, 3,
			[]int{
				1, 1, 0,
				1, 0, 1,
				0, 0, 0,
			},
			[]int{
				1, 1, 0,
				1, 0, 0,
				0, 0, 0,
			},
		),
		Entry("neighbours stop at the edge of the board",
			3, 1,
			[]int{1, 1, 1},
			[]int{0, 1, 0},
		),
	)
})
