package websocket

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"game_of_life/server/api/users"
	"game_of_life/server/internal/grid/mocks"
	"game_of_life/server/internal/grid/source"

	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
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

func testServer(t *testing.T, board *mocks.MockGrid, people users.Tracker) *httptest.Server {
	t.Helper()
	gin.SetMode(gin.TestMode)
	h := New(board, people)
	engine := gin.New()
	engine.GET("/ws", h.Serve)
	srv := httptest.NewServer(engine)
	t.Cleanup(srv.Close)
	return srv
}

func TestClientReceivesGrid(t *testing.T) {
	g := NewWithT(t)

	frame := fakeFrame{width: 2, height: 2, cells: []int{1, 0, 0, 1}}
	board := mocks.NewMockGrid(t)
	board.EXPECT().Current().Return(frame).Once()
	people := users.New()
	srv := testServer(t, board, people)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	t.Cleanup(cancel)

	conn, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(srv.URL, "http")+"/ws", nil)
	g.Expect(err).NotTo(HaveOccurred())
	t.Cleanup(func() { conn.CloseNow() })

	_, data, err := conn.Read(ctx)
	g.Expect(err).NotTo(HaveOccurred())

	want, err := frame.ToJson()
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(data).To(Equal(want))
	g.Expect(people.Conns()).To(HaveLen(1))
}

func TestRunForwardsBoardUpdates(t *testing.T) {
	g := NewWithT(t)

	first := fakeFrame{width: 2, height: 2, cells: []int{1, 0, 0, 0}}
	second := fakeFrame{width: 2, height: 2, cells: []int{0, 1, 0, 0}}
	secondJSON, err := second.ToJson()
	g.Expect(err).NotTo(HaveOccurred())

	updates := make(chan []byte, 1)
	var stream <-chan []byte = updates

	board := mocks.NewMockGrid(t)
	board.EXPECT().Current().Return(first).Once()
	board.EXPECT().Updates().Return(stream)

	people := users.New()
	h := New(board, people)
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.GET("/ws", h.Serve)
	srv := httptest.NewServer(engine)
	t.Cleanup(srv.Close)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go h.Run(ctx)

	dialCtx, dialCancel := context.WithTimeout(context.Background(), 2*time.Second)
	t.Cleanup(dialCancel)

	conn, _, err := websocket.Dial(dialCtx, "ws"+strings.TrimPrefix(srv.URL, "http")+"/ws", nil)
	g.Expect(err).NotTo(HaveOccurred())
	t.Cleanup(func() { conn.CloseNow() })

	_, data, err := conn.Read(dialCtx)
	g.Expect(err).NotTo(HaveOccurred())
	want, err := first.ToJson()
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(data).To(Equal(want))

	updates <- secondJSON

	_, data, err = conn.Read(dialCtx)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(data).To(Equal(secondJSON))
}
