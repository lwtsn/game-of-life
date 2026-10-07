package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"game_of_life/server/internal/grid/mocks"
	"game_of_life/server/internal/grid/source"
	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
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

func testServer(t *testing.T, board *mocks.MockGrid) (*hub, *httptest.Server) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	h := newHub(board)
	engine := gin.New()
	h.Register(engine)
	srv := httptest.NewServer(engine)
	t.Cleanup(srv.Close)
	return h, srv
}

func TestClientReceivesGrid(t *testing.T) {
	g := NewWithT(t)

	cells := []int{1, 0, 0, 1}
	frame := fakeFrame{width: 2, height: 2, cells: cells}

	board := mocks.NewMockGrid(t)
	board.EXPECT().Current().Return(frame).Once()

	h, srv := testServer(t, board)

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

	h.mu.Lock()
	tracked := h.users["127.0.0.1"]
	h.mu.Unlock()
	g.Expect(tracked).NotTo(BeNil())
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

	h, srv := testServer(t, board)
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

func TestStartCallsTheGrid(t *testing.T) {
	g := NewWithT(t)

	board := mocks.NewMockGrid(t)
	board.EXPECT().Start(mock.Anything).Once()

	_, srv := testServer(t, board)
	res, err := http.Post(srv.URL+"/start", "application/json", nil)
	g.Expect(err).NotTo(HaveOccurred())
	defer res.Body.Close()
	g.Expect(res.StatusCode).To(Equal(http.StatusNoContent))
}

func TestLayoutIsAStub(t *testing.T) {
	g := NewWithT(t)

	board := mocks.NewMockGrid(t)
	_, srv := testServer(t, board)

	res, err := http.Post(srv.URL+"/layout", "application/json", strings.NewReader(`{"name":"glider"}`))
	g.Expect(err).NotTo(HaveOccurred())
	defer res.Body.Close()
	g.Expect(res.StatusCode).To(Equal(http.StatusOK))

	body, err := io.ReadAll(res.Body)
	g.Expect(err).NotTo(HaveOccurred())

	var got map[string]any
	g.Expect(json.Unmarshal(body, &got)).To(Succeed())
	g.Expect(got["layout"]).To(Equal("glider"))
	g.Expect(got["applied"]).To(Equal(false))
}
