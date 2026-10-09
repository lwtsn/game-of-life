package websocket

import (
	"context"
	"errors"
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

// mailboxSize is how many messages a socket may fall behind before it is dropped.
// At the fastest pace that is well over half a second of boards.
const mailboxSize = 64

var (
	errGone = errors.New("socket has disconnected")
	errSlow = errors.New("socket is not keeping up")
)

// client is one socket. Only its pump goroutine writes to the network.
type client struct {
	person user.User
	out    chan []byte
}

type handler struct {
	grid    grid.Grid
	svc     user.Service
	origins Origins
	mu      sync.Mutex
	clients map[*websocket.Conn]*client

	// writeMu orders what goes into the mailboxes, so a stale board can never follow an edit.
	// It is only held while queueing, never during a network write.
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
		clients: make(map[*websocket.Conn]*client),
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
			slow := h.enqueueLocked(payload)
			h.writeMu.Unlock()
			if h.drop(slow) {
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

// writeAll queues the payload for every socket. It reports whether any socket was dropped.
func (h *handler) writeAll(payload []byte) bool {
	h.writeMu.Lock()
	slow := h.enqueueLocked(payload)
	h.writeMu.Unlock()
	return h.drop(slow)
}

// enqueueLocked puts the payload in every mailbox without waiting on the network.
// It returns the sockets whose mailbox is full. The caller holds writeMu.
func (h *handler) enqueueLocked(payload []byte) []*websocket.Conn {
	h.mu.Lock()
	defer h.mu.Unlock()
	var slow []*websocket.Conn
	for conn, c := range h.clients {
		select {
		case c.out <- payload:
		default:
			slow = append(slow, conn)
		}
	}
	return slow
}

// drop disconnects sockets that fell behind. They reconnect and get a full board.
func (h *handler) drop(slow []*websocket.Conn) bool {
	dropped := false
	for _, conn := range slow {
		if h.disconnect(conn) {
			dropped = true
		}
	}
	return dropped
}

// send queues a payload for one socket.
func (h *handler) send(conn *websocket.Conn, payload []byte) error {
	h.writeMu.Lock()
	defer h.writeMu.Unlock()
	h.mu.Lock()
	defer h.mu.Unlock()
	c, ok := h.clients[conn]
	if !ok {
		return errGone
	}
	select {
	case c.out <- payload:
		return nil
	default:
		return errSlow
	}
}
