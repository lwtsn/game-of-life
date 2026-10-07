package api

import (
	"context"
	"sync"

	"game_of_life/server/internal/grid"

	"github.com/coder/websocket"
)

type hub struct {
	grid grid.Grid

	mu      sync.Mutex
	writeMu sync.Mutex
	ctx     context.Context
	users   map[string]*websocket.Conn
}

func newHub(board grid.Grid) *hub {
	return &hub{
		grid:  board,
		users: make(map[string]*websocket.Conn),
	}
}

// Run writes each board update to the connected users until ctx is cancelled.
func (h *hub) Run(ctx context.Context) {
	h.mu.Lock()
	h.ctx = ctx
	h.mu.Unlock()

	for {
		select {
		case <-ctx.Done():
			return
		case payload := <-h.grid.Updates():
			h.writeAll(payload)
		}
	}
}

func (h *hub) writeAll(payload []byte) {
	h.mu.Lock()
	conns := make([]*websocket.Conn, 0, len(h.users))
	for _, conn := range h.users {
		conns = append(conns, conn)
	}
	h.mu.Unlock()

	for _, conn := range conns {
		if err := h.send(context.Background(), conn, payload); err != nil {
			h.disconnect(conn)
		}
	}
}

func (h *hub) send(ctx context.Context, conn *websocket.Conn, payload []byte) error {
	h.writeMu.Lock()
	defer h.writeMu.Unlock()
	return writeFrame(ctx, conn, payload)
}
