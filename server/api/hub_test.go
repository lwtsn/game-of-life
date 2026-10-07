package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"game_of_life/server/internal/grid/mocks"
	"game_of_life/server/internal/grid/source"
	"github.com/coder/websocket"
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

func TestClientReceivesGrid(t *testing.T) {
	g := NewWithT(t)

	cells := []int{1, 0, 0, 1}
	frame := fakeFrame{width: 2, height: 2, cells: cells}

	board := mocks.NewMockGrid(t)
	board.EXPECT().Current().Return(frame).Once()

	hub := newHub(board)
	srv := httptest.NewServer(http.HandlerFunc(hub.ServeHTTP))
	t.Cleanup(srv.Close)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	t.Cleanup(cancel)

	conn, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(srv.URL, "http"), nil)
	g.Expect(err).NotTo(HaveOccurred())
	t.Cleanup(func() { conn.CloseNow() })

	_, data, err := conn.Read(ctx)
	g.Expect(err).NotTo(HaveOccurred())

	want, err := frame.ToJson()
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(data).To(Equal(want))
}
