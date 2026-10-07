package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"game_of_life/server/internal/grid"
	"game_of_life/server/internal/grid/mocks"
	"github.com/coder/websocket"
	. "github.com/onsi/gomega"
)

func TestClientReceivesGrid(t *testing.T) {
	g := NewWithT(t)

	cells := make([]int, grid.Cols*grid.Rows)
	cells[0] = 1
	snap := grid.Snapshot{Width: grid.Cols, Height: grid.Rows, Cells: cells}

	source := mocks.NewMockSource(t)
	source.EXPECT().Next().Return(snap).Once()

	hub := NewHub(source)
	srv := httptest.NewServer(http.HandlerFunc(hub.ServeHTTP))
	t.Cleanup(srv.Close)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	t.Cleanup(cancel)

	conn, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(srv.URL, "http"), nil)
	g.Expect(err).NotTo(HaveOccurred())
	t.Cleanup(func() { conn.CloseNow() })

	_, data, err := conn.Read(ctx)
	g.Expect(err).NotTo(HaveOccurred())

	var got grid.Snapshot
	g.Expect(json.Unmarshal(data, &got)).To(Succeed())
	g.Expect(got).To(Equal(snap))
}
