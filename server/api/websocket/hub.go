package websocket

import (
	"context"
	"log"
	"sync"

	lifepb "game_of_life/server/gen/life/v1"
	"game_of_life/server/internal/grid"
	"game_of_life/server/internal/user"

	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"google.golang.org/protobuf/encoding/protojson"
)

// Handler is the board socket.
type Handler interface {
	Serve(*gin.Context)
	Run(context.Context)
	Publish()
}

type handler struct {
	grid    grid.Grid
	svc     user.Service
	origins Origins
	mu      sync.Mutex
	clients map[*websocket.Conn]user.User

	writeMu sync.Mutex
}

// Params are the handler's dependencies. Origins is optional and falls back to DefaultOrigins.
type Params struct {
	fx.In

	Grid    grid.Grid
	Users   user.Service
	Origins Origins `optional:"true"`
}

func New(p Params) Handler {
	return &handler{
		grid:    p.Grid,
		svc:     p.Users,
		origins: p.Origins,
		clients: make(map[*websocket.Conn]user.User),
	}
}

// Run writes each board update to the connected users until ctx is cancelled.
func (h *handler) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case payload := <-h.grid.Updates():
			h.writeMu.Lock()
			if !h.grid.Live() {
				h.writeMu.Unlock()
				continue
			}
			failed := h.writeLocked(payload)
			h.writeMu.Unlock()
			for _, conn := range failed {
				h.disconnect(conn)
			}
			if len(failed) > 0 {
				h.broadcastPeople()
			}
		}
	}
}

func (h *handler) broadcastPeople() {
	payload, err := h.peoplePayload()
	if err != nil {
		log.Printf("people json: %v", err)
		return
	}
	if h.writeAll(payload) {
		payload, err = h.peoplePayload()
		if err != nil {
			log.Printf("people json: %v", err)
			return
		}
		h.writeAll(payload)
	}
}

func (h *handler) peoplePayload() ([]byte, error) {
	return protojson.Marshal(&lifepb.ServerMessage{
		Type:   lifepb.MessageType_MESSAGE_TYPE_PEOPLE,
		People: h.svc.Colours(),
	})
}

// Publish writes the current board to every connected socket.
func (h *handler) Publish() {
	payload, err := h.grid.Current().ToJson()
	if err != nil {
		log.Printf("grid json: %v", err)
		return
	}
	h.writeAll(payload)
}

func (h *handler) writeAll(payload []byte) bool {
	h.writeMu.Lock()
	failed := h.writeLocked(payload)
	h.writeMu.Unlock()
	for _, conn := range failed {
		h.disconnect(conn)
	}
	return len(failed) > 0
}

func (h *handler) writeLocked(payload []byte) []*websocket.Conn {
	var failed []*websocket.Conn
	for _, conn := range h.conns() {
		if err := writeFrame(context.Background(), conn, payload); err != nil {
			failed = append(failed, conn)
		}
	}
	return failed
}

func (h *handler) send(ctx context.Context, conn *websocket.Conn, payload []byte) error {
	h.writeMu.Lock()
	defer h.writeMu.Unlock()
	return writeFrame(ctx, conn, payload)
}
