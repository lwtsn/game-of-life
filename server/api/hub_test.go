package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"game_of_life/server/internal/grid/mocks"
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

func TestClientReceivesGrid(t *testing.T) {
	g := NewWithT(t)

	cells := []int{1, 0, 0, 1}
	frame := fakeFrame{width: 2, height: 2, cells: cells}

	board := mocks.NewMockGrid(t)
	board.EXPECT().Next().Return(frame).Once()

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

	var got wire
	g.Expect(json.Unmarshal(data, &got)).To(Succeed())
	g.Expect(got).To(Equal(wire{Width: 2, Height: 2, Cells: cells}))
}
