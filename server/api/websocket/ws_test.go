package websocket

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"time"

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
	return source.Encode(f)
}

func openHandler(board grid.Grid) *handler {
	GinkgoHelper()
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

		conn, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(srv.URL, "http")+"/ws", nil)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(conn.CloseNow)

		data := readBoard(conn, ctx)
		want, err := frame.ToJson()
		Expect(err).NotTo(HaveOccurred())
		Expect(data).To(Equal(want))
		Expect(h.conns()).To(HaveLen(1))

		var presence struct {
			Colours []string `json:"colours"`
		}
		Expect(json.Unmarshal(readMessage(conn, ctx), &presence)).To(Succeed())
		Expect(presence.Colours).To(Equal([]string{h.svc.Colour("127.0.0.1")}))
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

	It("sends a board update to every connection from the same address", func() {
		first := fakeFrame{width: 2, height: 2, cells: []source.Cell{{Alive: true}, {}, {}, {}}}
		second := fakeFrame{width: 2, height: 2, cells: []source.Cell{{}, {Alive: true}, {}, {}}}
		secondJSON, err := second.ToJson()
		Expect(err).NotTo(HaveOccurred())

		updates := make(chan []byte, 1)
		var stream <-chan []byte = updates

		board := mocks.NewMockGrid(GinkgoT())
		board.EXPECT().Current().Return(first).Times(2)
		board.EXPECT().Updates().Return(stream)

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
			conn, _, err := websocket.Dial(dialCtx, "ws"+strings.TrimPrefix(srv.URL, "http")+"/ws", nil)
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
		person := user.New("127.0.0.1")
		placed := fakeFrame{width: 2, height: 2, cells: []source.Cell{
			{},
			{Alive: true, User: person},
			{},
			{},
		}}
		board := mocks.NewMockGrid(GinkgoT())
		board.EXPECT().Current().Return(frame).Once()
		board.EXPECT().Place(1, 0, mock.MatchedBy(func(got user.User) bool {
			return got != nil && got.IP() == person.IP() && got.Colour() == person.Colour()
		})).Return(true).Once()
		board.EXPECT().Current().Return(placed).Once()
		srv, _ := testServer(board)

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		DeferCleanup(cancel)

		conn, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(srv.URL, "http")+"/ws", nil)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(conn.CloseNow)

		want, err := frame.ToJson()
		Expect(err).NotTo(HaveOccurred())
		Expect(readBoard(conn, ctx)).To(Equal(want))

		Expect(conn.Write(ctx, websocket.MessageText, []byte(`{}`))).To(Succeed())
		Expect(conn.Write(ctx, websocket.MessageText, []byte(`{"x":1,"y":0}`))).To(Succeed())

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

})

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
		var body map[string]any
		Expect(json.Unmarshal(data, &body)).To(Succeed())
		if _, ok := body["colours"]; ok {
			if _, grid := body["width"]; !grid {
				continue
			}
		}
		return data
	}
}
