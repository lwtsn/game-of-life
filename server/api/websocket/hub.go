package websocket

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"game_of_life/server/internal/grid"
	"game_of_life/server/internal/user/service"

	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
)

// Handler is the board socket.
type Handler interface {
	Serve(*gin.Context)
	Run(context.Context)
}

type handler struct {
	grid    grid.Grid
	svc     service.Service
	mu      sync.Mutex
	clients map[string]*websocket.Conn

	writeMu sync.Mutex
}

func New(board grid.Grid, svc service.Service) Handler {
	return &handler{
		grid:    board,
		svc:     svc,
		clients: make(map[string]*websocket.Conn),
	}
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
	}{Colours: h.svc.Colours()})
}

func (h *handler) writeAll(payload []byte) bool {
	dropped := false
	for _, conn := range h.conns() {
		if err := h.send(context.Background(), conn, payload); err != nil {
			h.disconnect(conn)
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
