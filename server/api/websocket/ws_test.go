package websocket

import (
	"context"
	"net/http/httptest"
	"strings"
	"time"

	"game_of_life/server/api/users"
	"game_of_life/server/internal/grid/mocks"
	"game_of_life/server/internal/grid/source"

	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type fakeFrame struct {
	width  int
	height int
	cells  []int
}

func (f fakeFrame) Width() int   { return f.width }
func (f fakeFrame) Height() int  { return f.height }
func (f fakeFrame) Cells() []int { return f.cells }
func (f fakeFrame) ToJson() ([]byte, error) {
	return source.Encode(f)
}

func testServer(board *mocks.MockGrid, people users.Tracker) *httptest.Server {
	GinkgoHelper()
	h := New(board, people)
	engine := gin.New()
	engine.GET("/ws", h.Serve)
	srv := httptest.NewServer(engine)
	DeferCleanup(srv.Close)
	return srv
}

var _ = Describe("websocket", func() {
	It("sends the current board on connect", func() {
		frame := fakeFrame{width: 2, height: 2, cells: []int{1, 0, 0, 1}}
		board := mocks.NewMockGrid(GinkgoT())
		board.EXPECT().Current().Return(frame).Once()
		people := users.New()
		srv := testServer(board, people)

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		DeferCleanup(cancel)

		conn, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(srv.URL, "http")+"/ws", nil)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(conn.CloseNow)

		_, data, err := conn.Read(ctx)
		Expect(err).NotTo(HaveOccurred())

		want, err := frame.ToJson()
		Expect(err).NotTo(HaveOccurred())
		Expect(data).To(Equal(want))
		Expect(people.Conns()).To(HaveLen(1))
	})

	It("forwards later board updates", func() {
		first := fakeFrame{width: 2, height: 2, cells: []int{1, 0, 0, 0}}
		second := fakeFrame{width: 2, height: 2, cells: []int{0, 1, 0, 0}}
		secondJSON, err := second.ToJson()
		Expect(err).NotTo(HaveOccurred())

		updates := make(chan []byte, 1)
		var stream <-chan []byte = updates

		board := mocks.NewMockGrid(GinkgoT())
		board.EXPECT().Current().Return(first).Once()
		board.EXPECT().Updates().Return(stream)

		people := users.New()
		h := New(board, people)
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

		_, data, err := conn.Read(dialCtx)
		Expect(err).NotTo(HaveOccurred())
		want, err := first.ToJson()
		Expect(err).NotTo(HaveOccurred())
		Expect(data).To(Equal(want))

		updates <- secondJSON

		_, data, err = conn.Read(dialCtx)
		Expect(err).NotTo(HaveOccurred())
		Expect(data).To(Equal(secondJSON))
	})
})
