package websocket

import (
	"context"
	"encoding/json"
	"log"
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
			if h.writeAll(payload) {
				h.broadcastColours()
			}
		}
	}
}

func (h *handler) broadcastColours() {
	payload, err := h.colourPayload()
	if err != nil {
		log.Printf("colours json: %v", err)
		return
	}
	if h.writeAll(payload) {
		payload, err = h.colourPayload()
		if err != nil {
			log.Printf("colours json: %v", err)
			return
		}
		h.writeAll(payload)
	}
}

func (h *handler) colourPayload() ([]byte, error) {
	return json.Marshal(struct {
		Colours []string `json:"colours"`
	}{Colours: h.people.Colours()})
}

func (h *handler) writeAll(payload []byte) bool {
	dropped := false
	for _, conn := range h.people.Conns() {
		if err := h.send(context.Background(), conn, payload); err != nil {
			h.people.Disconnect(conn)
			dropped = true
		}
	}
	return dropped
}

func (h *handler) send(ctx context.Context, conn *websocket.Conn, payload []byte) error {
	h.writeMu.Lock()
	defer h.writeMu.Unlock()
	return writeFrame(ctx, conn, payload)
}
