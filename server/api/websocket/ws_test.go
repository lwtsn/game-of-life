package websocket

import (
	"context"
	"net/http/httptest"
	"strings"
	"time"

	lifepb "game_of_life/server/gen/life/v1"
	"game_of_life/server/internal/grid"
	"game_of_life/server/internal/grid/mocks"
	"game_of_life/server/internal/grid/source"
	"game_of_life/server/internal/user"

	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	"go.uber.org/fx"
	"google.golang.org/protobuf/encoding/protojson"
)

type fakeFrame struct {
	width  int
	height int
	cells  []source.Cell
}

func (f fakeFrame) Width() int           { return f.width }
func (f fakeFrame) Height() int          { return f.height }
func (f fakeFrame) Cells() []source.Cell { return f.cells }
func (f fakeFrame) ToJson() ([]byte, error) {
	return source.Encode(f, 0)
}

func openHandler(board grid.Grid) *handler {
	GinkgoHelper()
	if mocked, ok := board.(*mocks.MockGrid); ok {
		ensureClock(mocked)
	}
	var api Handler
	app := fx.New(
		user.Module,
		Module,
		fx.Provide(func() grid.Grid { return board }),
		fx.Populate(&api),
		fx.NopLogger,
	)
	Expect(app.Err()).NotTo(HaveOccurred())
	return api.(*handler)
}

func ensureClock(board *mocks.MockGrid) {
	for _, call := range board.ExpectedCalls {
		if call.Method == "Clock" {
			return
		}
	}
	board.EXPECT().Clock().Return(true, int(lifepb.PaceBound_PACE_BOUND_MIN)).Maybe()
}

func testServer(board *mocks.MockGrid) (*httptest.Server, *handler) {
	GinkgoHelper()
	h := openHandler(board)
	engine := gin.New()
	engine.GET("/ws", h.Serve)
	srv := httptest.NewServer(engine)
	DeferCleanup(srv.Close)
	return srv, h
}

var _ = Describe("websocket", func() {
	It("sends the current board on connect", func() {
		frame := fakeFrame{width: 2, height: 2, cells: []source.Cell{{Alive: true}, {}, {}, {Alive: true}}}
		board := mocks.NewMockGrid(GinkgoT())
		board.EXPECT().Current().Return(frame).Once()
		srv, h := testServer(board)

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		DeferCleanup(cancel)

		conn, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(srv.URL, "http")+"/ws?session=player-one", nil)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(conn.CloseNow)

		data := readBoard(conn, ctx)
		want, err := frame.ToJson()
		Expect(err).NotTo(HaveOccurred())
		Expect(data).To(Equal(want))
		Expect(h.conns()).To(HaveLen(1))

		var presence lifepb.ServerMessage
		Expect(protojson.Unmarshal(readMessage(conn, ctx), &presence)).To(Succeed())
		person, ok := h.svc.ByID("player-one")
		Expect(ok).To(BeTrue())
		Expect(presence.GetType()).To(Equal(lifepb.MessageType_MESSAGE_TYPE_PEOPLE))
		Expect(presence.GetPeople()).To(Equal([]string{person.Colour()}))
	})

	It("forwards later board updates", func() {
		first := fakeFrame{width: 2, height: 2, cells: []source.Cell{{Alive: true}, {}, {}, {}}}
		second := fakeFrame{width: 2, height: 2, cells: []source.Cell{{}, {Alive: true}, {}, {}}}
		secondJSON, err := second.ToJson()
		Expect(err).NotTo(HaveOccurred())

		updates := make(chan []byte, 1)
		var stream <-chan []byte = updates

		board := mocks.NewMockGrid(GinkgoT())
		board.EXPECT().Current().Return(first).Once()
		board.EXPECT().Updates().Return(stream)
		board.EXPECT().Live().Return(true).Maybe()

		h := openHandler(board)
		engine := gin.New()
		engine.GET("/ws", h.Serve)
		srv := httptest.NewServer(engine)
		DeferCleanup(srv.Close)

		ctx, cancel := context.WithCancel(context.Background())
		DeferCleanup(cancel)
		go h.Run(ctx)

		dialCtx, dialCancel := context.WithTimeout(context.Background(), 2*time.Second)
		DeferCleanup(dialCancel)

		conn, _, err := websocket.Dial(dialCtx, "ws"+strings.TrimPrefix(srv.URL, "http")+"/ws", nil)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(conn.CloseNow)

		data := readBoard(conn, dialCtx)
		want, err := first.ToJson()
		Expect(err).NotTo(HaveOccurred())
		Expect(data).To(Equal(want))

		updates <- secondJSON

		Expect(readBoard(conn, dialCtx)).To(Equal(secondJSON))
	})

	It("sends a board update to every connection for the session", func() {
		first := fakeFrame{width: 2, height: 2, cells: []source.Cell{{Alive: true}, {}, {}, {}}}
		second := fakeFrame{width: 2, height: 2, cells: []source.Cell{{}, {Alive: true}, {}, {}}}
		secondJSON, err := second.ToJson()
		Expect(err).NotTo(HaveOccurred())

		updates := make(chan []byte, 1)
		var stream <-chan []byte = updates

		board := mocks.NewMockGrid(GinkgoT())
		board.EXPECT().Current().Return(first).Times(2)
		board.EXPECT().Updates().Return(stream)
		board.EXPECT().Live().Return(true).Maybe()

		h := openHandler(board)
		engine := gin.New()
		engine.GET("/ws", h.Serve)
		srv := httptest.NewServer(engine)
		DeferCleanup(srv.Close)

		ctx, cancel := context.WithCancel(context.Background())
		DeferCleanup(cancel)
		go h.Run(ctx)

		dialCtx, dialCancel := context.WithTimeout(context.Background(), 2*time.Second)
		DeferCleanup(dialCancel)

		dial := func() *websocket.Conn {
			GinkgoHelper()
			conn, _, err := websocket.Dial(dialCtx, "ws"+strings.TrimPrefix(srv.URL, "http")+"/ws?session=player-one", nil)
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(func() { conn.CloseNow() })
			return conn
		}

		left := dial()
		right := dial()
		want, err := first.ToJson()
		Expect(err).NotTo(HaveOccurred())
		Expect(readBoard(left, dialCtx)).To(Equal(want))
		Expect(readBoard(right, dialCtx)).To(Equal(want))

		updates <- secondJSON

		Expect(readBoard(left, dialCtx)).To(Equal(secondJSON))
		Expect(readBoard(right, dialCtx)).To(Equal(secondJSON))
	})

	It("places the cell the connection asked for", func() {
		frame := fakeFrame{width: 2, height: 2, cells: make([]source.Cell, 4)}
		person := user.New("player-one")
		placed := fakeFrame{width: 2, height: 2, cells: []source.Cell{
			{},
			{Alive: true, User: person},
			{},
			{},
		}}
		board := mocks.NewMockGrid(GinkgoT())
		board.EXPECT().Current().Return(frame).Once()
		board.EXPECT().Place(1, 0, mock.MatchedBy(func(got user.User) bool {
			return got != nil && got.ID() == "player-one" && got.Colour() == "#112D4E"
		})).Return(true).Once()
		board.EXPECT().Current().Return(placed).Once()
		srv, _ := testServer(board)

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		DeferCleanup(cancel)

		conn, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(srv.URL, "http")+"/ws?session=player-one", nil)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(conn.CloseNow)

		want, err := frame.ToJson()
		Expect(err).NotTo(HaveOccurred())
		Expect(readBoard(conn, ctx)).To(Equal(want))

		writeClient(conn, ctx, &lifepb.ClientMessage{})
		writeClient(conn, ctx, &lifepb.ClientMessage{
			Action: &lifepb.ClientMessage_Point{Point: &lifepb.Point{X: 1}},
		})

		placedJSON, err := placed.ToJson()
		Expect(err).NotTo(HaveOccurred())
		Expect(readBoard(conn, ctx)).To(Equal(placedJSON))
	})

	It("publishes the current board to connected sockets", func() {
		frame := fakeFrame{width: 2, height: 2, cells: []source.Cell{{Alive: true}, {}, {}, {}}}
		board := mocks.NewMockGrid(GinkgoT())
		board.EXPECT().Current().Return(frame).Times(2)
		srv, h := testServer(board)

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		DeferCleanup(cancel)

		conn, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(srv.URL, "http")+"/ws", nil)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(conn.CloseNow)

		want, err := frame.ToJson()
		Expect(err).NotTo(HaveOccurred())
		Expect(readBoard(conn, ctx)).To(Equal(want))

		h.Publish()
		Expect(readBoard(conn, ctx)).To(Equal(want))
	})

	It("clears the board and names the colour that asked", func() {
		frame := fakeFrame{width: 2, height: 2, cells: []source.Cell{{Alive: true}, {}, {}, {}}}
		cleared := fakeFrame{width: 2, height: 2, cells: make([]source.Cell, 4)}
		board := mocks.NewMockGrid(GinkgoT())
		board.EXPECT().Current().Return(frame).Once()
		board.EXPECT().SetRunning(false).Once()
		board.EXPECT().Clear().Return(true).Once()
		board.EXPECT().Current().Return(cleared).Once()
		board.EXPECT().Clock().Return(true, int(lifepb.PaceBound_PACE_BOUND_MIN)).Once()
		board.EXPECT().Clock().Return(false, int(lifepb.PaceBound_PACE_BOUND_MIN)).Once()
		srv, _ := testServer(board)

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		DeferCleanup(cancel)

		conn, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(srv.URL, "http")+"/ws?session=player-one", nil)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(conn.CloseNow)

		want, err := frame.ToJson()
		Expect(err).NotTo(HaveOccurred())
		Expect(readBoard(conn, ctx)).To(Equal(want))

		writeClient(conn, ctx, &lifepb.ClientMessage{
			Action: &lifepb.ClientMessage_ResetBoard{ResetBoard: true},
		})

		clearedJSON, err := cleared.ToJson()
		Expect(err).NotTo(HaveOccurred())
		Expect(readBoard(conn, ctx)).To(Equal(clearedJSON))

		var notice lifepb.ServerMessage
		Expect(protojson.Unmarshal(readMessage(conn, ctx), &notice)).To(Succeed())
		Expect(notice.GetType()).To(Equal(lifepb.MessageType_MESSAGE_TYPE_RESET))
		Expect(notice.GetColour()).To(Equal("#112D4E"))
		Expect(readClock(conn, ctx)).To(Equal(clockState{running: false, pace: int(lifepb.PaceBound_PACE_BOUND_MIN)}))
	})

	It("sets the colour the session asked for", func() {
		frame := fakeFrame{width: 2, height: 2, cells: make([]source.Cell, 4)}
		board := mocks.NewMockGrid(GinkgoT())
		board.EXPECT().Current().Return(frame).Once()
		board.EXPECT().Place(1, 0, mock.MatchedBy(func(got user.User) bool {
			return got != nil && got.ID() == "player-one" && got.Colour() == "#E58700"
		})).Return(true).Once()
		board.EXPECT().Current().Return(frame).Once()
		srv, h := testServer(board)

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		DeferCleanup(cancel)

		conn, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(srv.URL, "http")+"/ws?session=player-one", nil)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(conn.CloseNow)
		_ = readBoard(conn, ctx)

		writeClient(conn, ctx, &lifepb.ClientMessage{
			Action: &lifepb.ClientMessage_Colour{Colour: "#E58700"},
		})

		Eventually(func() string {
			var you lifepb.ServerMessage
			Expect(protojson.Unmarshal(readMessage(conn, ctx), &you)).To(Succeed())
			if you.GetType() != lifepb.MessageType_MESSAGE_TYPE_YOU {
				return ""
			}
			return you.GetColour()
		}, time.Second, 10*time.Millisecond).Should(Equal("#E58700"))

		Expect(readNotice(conn, ctx, lifepb.MessageType_MESSAGE_TYPE_CHANGED)).To(Equal("#E58700"))

		person, ok := h.svc.ByID("player-one")
		Expect(ok).To(BeTrue())
		Expect(person.Colour()).To(Equal("#E58700"))
		Expect(h.svc.Colours()).To(Equal([]string{"#E58700"}))

		writeClient(conn, ctx, &lifepb.ClientMessage{
			Action: &lifepb.ClientMessage_Point{Point: &lifepb.Point{X: 1}},
		})
		Expect(readBoard(conn, ctx)).NotTo(BeEmpty())
	})

	It("announces a session on its first socket and when its last socket closes", func() {
		frame := fakeFrame{width: 1, height: 1, cells: []source.Cell{{}}}
		board := mocks.NewMockGrid(GinkgoT())
		board.EXPECT().Current().Return(frame)
		srv, h := testServer(board)

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		DeferCleanup(cancel)

		first, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(srv.URL, "http")+"/ws?session=player-one", nil)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(first.CloseNow)
		Expect(readNotice(first, ctx, lifepb.MessageType_MESSAGE_TYPE_ENTERED)).To(Equal("#112D4E"))

		second, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(srv.URL, "http")+"/ws?session=player-two", nil)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { _ = second.CloseNow() })
		Expect(readNotice(first, ctx, lifepb.MessageType_MESSAGE_TYPE_ENTERED)).To(Equal("#3F72AF"))
		Expect(readNotice(second, ctx, lifepb.MessageType_MESSAGE_TYPE_ENTERED)).To(Equal("#3F72AF"))

		extra, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(srv.URL, "http")+"/ws?session=player-one", nil)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(extra.CloseNow)
		sawEntered := false
		for {
			var body lifepb.ServerMessage
			Expect(protojson.Unmarshal(readMessage(first, ctx), &body)).To(Succeed())
			if body.GetType() == lifepb.MessageType_MESSAGE_TYPE_ENTERED {
				sawEntered = true
			}
			if body.GetType() == lifepb.MessageType_MESSAGE_TYPE_YOU {
				break
			}
		}
		Expect(sawEntered).To(BeFalse())

		writeClient(second, ctx, &lifepb.ClientMessage{
			Action: &lifepb.ClientMessage_Colour{Colour: "#E58700"},
		})
		Eventually(func() string {
			person, ok := h.svc.ByID("player-two")
			if !ok {
				return ""
			}
			return person.Colour()
		}, time.Second, 10*time.Millisecond).Should(Equal("#E58700"))
		Expect(second.Close(websocket.StatusNormalClosure, "")).To(Succeed())
		Expect(readNotice(first, ctx, lifepb.MessageType_MESSAGE_TYPE_EXITED)).To(Equal("#E58700"))
	})

	It("lets a session change the pace and stop the clock", func() {
		frame := fakeFrame{width: 1, height: 1, cells: []source.Cell{{}}}
		board := mocks.NewMockGrid(GinkgoT())
		board.EXPECT().Current().Return(frame).Once()
		board.EXPECT().Clock().Return(true, int(lifepb.PaceBound_PACE_BOUND_MIN)).Once()
		board.EXPECT().SetPace(20).Return(true).Once()
		board.EXPECT().Clock().Return(true, 20).Once()
		board.EXPECT().SetRunning(false).Once()
		board.EXPECT().Clock().Return(false, 20).Once()
		srv, _ := testServer(board)

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		DeferCleanup(cancel)

		conn, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(srv.URL, "http")+"/ws?session=player-one", nil)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(conn.CloseNow)
		Expect(readClock(conn, ctx)).To(Equal(clockState{running: true, pace: int(lifepb.PaceBound_PACE_BOUND_MIN)}))

		writeClient(conn, ctx, &lifepb.ClientMessage{
			Action: &lifepb.ClientMessage_Pace{Pace: 20},
		})
		Expect(readClock(conn, ctx)).To(Equal(clockState{running: true, pace: 20}))

		writeClient(conn, ctx, &lifepb.ClientMessage{
			Action: &lifepb.ClientMessage_Playback{Playback: lifepb.Playback_PLAYBACK_STOPPED},
		})
		Expect(readClock(conn, ctx)).To(Equal(clockState{running: false, pace: 20}))
	})
})

type clockState struct {
	running bool
	pace    int
}

func readClock(conn *websocket.Conn, ctx context.Context) clockState {
	GinkgoHelper()
	for {
		var body lifepb.ServerMessage
		Expect(protojson.Unmarshal(readMessage(conn, ctx), &body)).To(Succeed())
		if body.GetType() != lifepb.MessageType_MESSAGE_TYPE_CLOCK {
			continue
		}
		return clockState{
			running: body.GetPlayback() == lifepb.Playback_PLAYBACK_RUNNING,
			pace:    int(body.GetPace()),
		}
	}
}

func writeClient(conn *websocket.Conn, ctx context.Context, msg *lifepb.ClientMessage) {
	GinkgoHelper()
	payload, err := protojson.Marshal(msg)
	Expect(err).NotTo(HaveOccurred())
	Expect(conn.Write(ctx, websocket.MessageText, payload)).To(Succeed())
}

func readNotice(conn *websocket.Conn, ctx context.Context, kind lifepb.MessageType) string {
	GinkgoHelper()
	for {
		var body lifepb.ServerMessage
		Expect(protojson.Unmarshal(readMessage(conn, ctx), &body)).To(Succeed())
		if body.GetType() == kind {
			return body.GetColour()
		}
	}
}

func readMessage(conn *websocket.Conn, ctx context.Context) []byte {
	GinkgoHelper()
	_, data, err := conn.Read(ctx)
	Expect(err).NotTo(HaveOccurred())
	return data
}

func readBoard(conn *websocket.Conn, ctx context.Context) []byte {
	GinkgoHelper()
	for {
		data := readMessage(conn, ctx)
		var body lifepb.ServerMessage
		Expect(protojson.Unmarshal(data, &body)).To(Succeed())
		if body.GetWidth() != 0 {
			return data
		}
	}
}
