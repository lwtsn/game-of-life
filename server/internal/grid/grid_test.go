package grid

import (
	"context"
	"time"

	"game_of_life/server/internal/grid/source"
	"game_of_life/server/internal/grid/source/mocks"
	"game_of_life/server/internal/user"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	"go.uber.org/fx"
)

func gridFrom(src source.Source) Grid {
	GinkgoHelper()
	var board Grid
	app := fx.New(
		Module,
		fx.Provide(func() source.Source { return src }),
		fx.Populate(&board),
		fx.NopLogger,
	)
	Expect(app.Err()).NotTo(HaveOccurred())
	return board
}

var _ = Describe("Grid", func() {
	It("stores a copy of the current cells", func() {
		cells := []source.Cell{{Alive: true}, {}, {Alive: true}, {}}
		frame := snapshot{width: 2, height: 2, cells: cells}
		src := mocks.NewMockSource(GinkgoT())
		src.EXPECT().Next(nil).Return(frame).Once()

		board := gridFrom(src)
		cells[0].Alive = false

		got := board.Current()
		Expect(got.Width()).To(Equal(2))
		Expect(got.Height()).To(Equal(2))
		Expect(got.Cells()).To(Equal([]source.Cell{{Alive: true}, {}, {Alive: true}, {}}))
	})

	It("passes the stored board to the source", func() {
		first := snapshot{width: 2, height: 2, cells: []source.Cell{{Alive: true}, {}, {}, {}}}
		second := snapshot{width: 2, height: 2, cells: []source.Cell{{}, {Alive: true}, {}, {}}}
		src := mocks.NewMockSource(GinkgoT())
		src.EXPECT().Next(nil).Return(first).Once()
		src.EXPECT().Next(mock.MatchedBy(func(current source.Frame) bool {
			cells := current.Cells()
			return len(cells) == 4 && cells[0].Alive && !cells[1].Alive
		})).Return(second).Once()

		board := gridFrom(src)
		board.Advance()

		Expect(board.Current().Cells()).To(Equal([]source.Cell{{}, {Alive: true}, {}, {}}))
	})

	It("places the person on one cell", func() {
		frame := snapshot{width: 2, height: 2, cells: make([]source.Cell, 4)}
		src := mocks.NewMockSource(GinkgoT())
		src.EXPECT().Next(nil).Return(frame).Once()

		board := gridFrom(src)
		person := user.New("198.51.100.10")
		Expect(board.Place(1, 0, person)).To(BeTrue())
		Expect(board.Place(-1, 0, person)).To(BeFalse())
		Expect(board.Place(0, 0, nil)).To(BeFalse())

		got := board.Current().Cells()
		Expect(got[1]).To(Equal(source.Cell{Alive: true, User: person}))
		Expect(got[0].Alive).To(BeFalse())

		payload, err := board.Current().ToJson()
		Expect(err).NotTo(HaveOccurred())
		Expect(string(payload)).To(ContainSubstring(person.Colour()))
		Expect(string(payload)).To(ContainSubstring(person.IP()))
	})

	It("publishes the current board when started", func() {
		frame := snapshot{width: 2, height: 2, cells: []source.Cell{{Alive: true}, {}, {}, {Alive: true}}}
		src := mocks.NewMockSource(GinkgoT())
		src.EXPECT().Next(nil).Return(frame).Once()

		board := gridFrom(src)
		ctx, cancel := context.WithCancel(context.Background())
		DeferCleanup(cancel)
		board.Start(ctx)

		want, err := frame.ToJson()
		Expect(err).NotTo(HaveOccurred())
		Eventually(board.Updates(), time.Second, 10*time.Millisecond).Should(Receive(Equal(want)))
	})
})
