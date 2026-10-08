package life

import (
	"encoding/json"

	"game_of_life/server/internal/grid/source"
	"game_of_life/server/internal/user"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
)

func occupy(bits []int) []source.Cell {
	cells := make([]source.Cell, len(bits))
	for i, bit := range bits {
		if bit == 1 {
			cells[i] = source.Cell{Alive: true}
		}
	}
	return cells
}

var _ = Describe("life", func() {
	var src source.Source

	BeforeEach(func() {
		app := fx.New(Module, fx.Populate(&src), fx.NopLogger)
		Expect(app.Err()).NotTo(HaveOccurred())
	})

	It("starts from an empty 80 by 50 board", func() {
		got := src.Next(nil)
		Expect(got.Width()).To(Equal(80))
		Expect(got.Height()).To(Equal(50))
		Expect(got.Cells()).To(Equal(make([]source.Cell, 80*50)))
	})

	DescribeTable("the next generation",
		func(width, height int, cells, want []int) {
			occupied := occupy(cells)
			current := snapshot{width: width, height: height, cells: occupied}
			got := src.Next(current)

			Expect(current.Cells()).To(Equal(occupied))
			Expect(got.Width()).To(Equal(width))
			Expect(got.Height()).To(Equal(height))
			Expect(got.Cells()).To(Equal(occupy(want)))
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

	It("keeps the user on a cell that survives", func() {
		person := user.New("198.51.100.10")
		cells := occupy([]int{
			1, 1,
			1, 1,
		})
		cells[0].User = person
		cells[0].Colour = "#FF0000"
		current := snapshot{width: 2, height: 2, cells: cells}

		got := src.Next(current)

		Expect(got.Cells()[0].Alive).To(BeTrue())
		Expect(got.Cells()[0].User).To(Equal(person))
		Expect(got.Cells()[0].User.ID()).To(Equal("198.51.100.10"))
		Expect(got.Cells()[0].Colour).To(Equal("#FF0000"))
	})

	It("leaves a birth without a user", func() {
		person := user.New("198.51.100.10")
		cells := occupy([]int{
			1, 1, 0,
			1, 0, 0,
			0, 0, 0,
		})
		cells[0].User = person
		cells[0].Colour = "#FF0000"
		cells[1].Colour = "#00FF00"
		cells[3].Colour = "#0000FF"
		current := snapshot{width: 3, height: 3, cells: cells}

		got := src.Next(current)

		birth := got.Cells()[4]
		Expect(birth.Alive).To(BeTrue())
		Expect(birth.User).To(BeNil())
		Expect(birth.Colour).To(Equal("#869290"))
		Expect(got.Cells()[0].User).To(Equal(person))
		Expect(got.Cells()[0].Colour).To(Equal("#FF0000"))
		Expect(got.Cells()[1].Colour).To(Equal("#00FF00"))

		raw, err := json.Marshal(birth)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(raw)).To(Equal(`{"alive":true,"colour":"#869290"}`))
	})

	It("gives a birth the colour shared by its parents", func() {
		person := user.New("198.51.100.10")
		cells := occupy([]int{
			0, 0, 0,
			1, 1, 1,
			0, 0, 0,
		})
		for _, index := range []int{3, 4, 5} {
			cells[index].User = person
			cells[index].Colour = person.Colour()
		}
		current := snapshot{width: 3, height: 3, cells: cells}

		got := src.Next(current)

		for _, index := range []int{1, 7} {
			birth := got.Cells()[index]
			Expect(birth.Alive).To(BeTrue())
			Expect(birth.User).To(BeNil())
			Expect(birth.Colour).To(Equal(person.Colour()))
		}
		Expect(got.Cells()[4].User).To(Equal(person))
		Expect(got.Cells()[4].Colour).To(Equal(person.Colour()))
	})
})
