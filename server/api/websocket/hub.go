package websocket

import (
	"context"
	"sync"

	"game_of_life/server/api/users"
	"game_of_life/server/internal/grid"

	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
)

// Handler is the board socket.
type Handler interface {
	Serve(*gin.Context)
	Run(context.Context)
}

type handler struct {
	grid   grid.Grid
	people users.Tracker

	writeMu sync.Mutex
}

func New(board grid.Grid, people users.Tracker) Handler {
	return &handler{grid: board, people: people}
}

// Run writes each board update to the connected users until ctx is cancelled.
func (h *handler) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case payload := <-h.grid.Updates():
			h.writeAll(payload)
		}
	}
}

func (h *handler) writeAll(payload []byte) {
	for _, conn := range h.people.Conns() {
		if err := h.send(context.Background(), conn, payload); err != nil {
			h.people.Disconnect(conn)
		}
	}
}

func (h *handler) send(ctx context.Context, conn *websocket.Conn, payload []byte) error {
	h.writeMu.Lock()
	defer h.writeMu.Unlock()
	return writeFrame(ctx, conn, payload)
}
