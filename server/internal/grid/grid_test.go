package grid

import (
	"context"
	"time"

	lifepb "game_of_life/server/gen/life/v1"
	"game_of_life/server/internal/grid/source"
	"game_of_life/server/internal/grid/source/mocks"
	"game_of_life/server/internal/user"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	"go.uber.org/fx"
	"google.golang.org/protobuf/encoding/protojson"
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
		Expect(got[1]).To(Equal(source.Cell{Alive: true, User: person, Colour: person.Colour()}))
		Expect(got[0].Alive).To(BeFalse())

		payload, err := board.Current().ToJson()
		Expect(err).NotTo(HaveOccurred())
		Expect(string(payload)).To(ContainSubstring(person.Colour()))
		Expect(string(payload)).To(ContainSubstring(person.ID()))
	})

	It("places the person on every cell of a shape", func() {
		frame := snapshot{width: 3, height: 3, cells: make([]source.Cell, 9)}
		src := mocks.NewMockSource(GinkgoT())
		src.EXPECT().Next(nil).Return(frame).Once()

		board := gridFrom(src)
		person := user.New("198.51.100.10")
		points := []Point{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 0, Y: 1}, {X: 1, Y: 1}}
		Expect(board.PlaceAll(points, person)).To(BeTrue())
		Expect(board.PlaceAll([]Point{{X: 0, Y: 0}, {X: 3, Y: 0}}, person)).To(BeFalse())
		Expect(board.PlaceAll(nil, person)).To(BeFalse())
		Expect(board.PlaceAll([]Point{{X: 0, Y: 0}}, nil)).To(BeFalse())

		got := board.Current().Cells()
		placed := source.Cell{Alive: true, User: person, Colour: person.Colour()}
		Expect(got[0]).To(Equal(placed))
		Expect(got[1]).To(Equal(placed))
		Expect(got[3]).To(Equal(placed))
		Expect(got[4]).To(Equal(placed))
		Expect(got[2].Alive).To(BeFalse())
	})

	It("clears every cell and keeps the size", func() {
		frame := snapshot{width: 2, height: 2, cells: []source.Cell{{Alive: true}, {Alive: true}, {}, {Alive: true}}}
		src := mocks.NewMockSource(GinkgoT())
		src.EXPECT().Next(nil).Return(frame).Once()

		board := gridFrom(src)
		Expect(board.Clear()).To(BeTrue())

		got := board.Current()
		Expect(got.Width()).To(Equal(2))
		Expect(got.Height()).To(Equal(2))
		Expect(got.Cells()).To(Equal(make([]source.Cell, 4)))
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

	It("starts at one generation per second and keeps that pace", func() {
		frame := snapshot{width: 1, height: 1, cells: []source.Cell{{}}}
		src := mocks.NewMockSource(GinkgoT())
		src.EXPECT().Next(nil).Return(frame).Once()

		board := gridFrom(src)
		running, pace := board.Clock()
		Expect(running).To(BeTrue())
		Expect(pace).To(Equal(int(lifepb.PaceBound_PACE_BOUND_MIN)))
		Expect(board.SetPace(0)).To(BeFalse())
		Expect(board.SetPace(int(lifepb.PaceBound_PACE_BOUND_MAX) + 1)).To(BeFalse())
		Expect(board.SetPace(20)).To(BeTrue())
		running, pace = board.Clock()
		Expect(running).To(BeTrue())
		Expect(pace).To(Equal(20))
	})

	It("discards a queued board when stopped", func() {
		frame := snapshot{width: 1, height: 1, cells: []source.Cell{{}}}
		src := mocks.NewMockSource(GinkgoT())
		src.EXPECT().Next(mock.Anything).Return(frame)

		board := gridFrom(src)
		Expect(board.SetPace(int(lifepb.PaceBound_PACE_BOUND_MAX))).To(BeTrue())
		ctx, cancel := context.WithCancel(context.Background())
		DeferCleanup(cancel)
		board.Start(ctx)

		Eventually(board.Updates(), time.Second, time.Millisecond).Should(Receive())
		board.SetRunning(false)
		Consistently(board.Updates(), 200*time.Millisecond, 5*time.Millisecond).ShouldNot(Receive())
	})

	It("drops a queued generation when a cell is placed", func() {
		frame := snapshot{width: 1, height: 1, cells: []source.Cell{{}}}
		src := mocks.NewMockSource(GinkgoT())
		src.EXPECT().Next(nil).Return(frame).Once()

		board := newGrid(src)
		payload, err := board.Current().ToJson()
		Expect(err).NotTo(HaveOccurred())
		board.updates <- payload
		board.queuedSeq = board.seq

		Expect(board.Place(0, 0, user.New("player-one"))).To(BeTrue())
		Expect(board.Live()).To(BeFalse())
		Consistently(board.Updates(), 40*time.Millisecond, 5*time.Millisecond).ShouldNot(Receive())
	})

	It("does not step while stopped", func() {
		frame := snapshot{width: 1, height: 1, cells: []source.Cell{{}}}
		src := mocks.NewMockSource(GinkgoT())
		src.EXPECT().Next(nil).Return(frame).Once()

		board := gridFrom(src)
		board.SetRunning(false)
		ctx, cancel := context.WithCancel(context.Background())
		DeferCleanup(cancel)
		board.Start(ctx)

		Eventually(board.Updates(), time.Second, 10*time.Millisecond).Should(Receive())
		Consistently(board.Updates(), 300*time.Millisecond, 10*time.Millisecond).ShouldNot(Receive())
	})

	It("steps at the pace it was given", func() {
		frame := snapshot{width: 1, height: 1, cells: []source.Cell{{}}}
		src := mocks.NewMockSource(GinkgoT())
		src.EXPECT().Next(mock.Anything).Return(frame)

		board := gridFrom(src)
		Expect(board.SetPace(20)).To(BeTrue())
		ctx, cancel := context.WithCancel(context.Background())
		DeferCleanup(cancel)
		board.Start(ctx)

		n := 0
		Eventually(func() int {
			select {
			case <-board.Updates():
				n++
			default:
			}
			return n
		}, time.Second, 5*time.Millisecond).Should(BeNumerically(">=", 5))
	})

	It("numbers each generation in order", func() {
		frame := snapshot{width: 1, height: 1, cells: []source.Cell{{}}}
		src := mocks.NewMockSource(GinkgoT())
		src.EXPECT().Next(mock.Anything).Return(frame)

		board := gridFrom(src)
		Expect(board.SetPace(20)).To(BeTrue())
		ctx, cancel := context.WithCancel(context.Background())
		DeferCleanup(cancel)
		board.Start(ctx)

		n := 0
		var prev int32
		Eventually(func() int {
			select {
			case payload := <-board.Updates():
				var msg lifepb.ServerMessage
				Expect(protojson.Unmarshal(payload, &msg)).To(Succeed())
				if msg.GetFrame() == 0 {
					return n
				}
				Expect(msg.GetFrame()).To(Equal(prev + 1))
				prev = msg.GetFrame()
				n++
			default:
			}
			return n
		}, time.Second, 5*time.Millisecond).Should(BeNumerically(">=", 3))
	})
})
