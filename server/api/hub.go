package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"game_of_life/server/internal/grid"
	"github.com/coder/websocket"
)

// Handler serves GET /ws and pushes a frame to every client. hub is that handler.
type Handler interface {
	http.Handler
	Run(context.Context)
}

// wire is the provisional JSON frame sent on the socket.
type wire struct {
	Width  int   `json:"width"`
	Height int   `json:"height"`
	Cells  []int `json:"cells"`
}

type hub struct {
	grid grid.Grid

	mu    sync.Mutex
	conns map[*websocket.Conn]struct{}
	last  []byte
}

func newHub(board grid.Grid) *hub {
	return &hub{
		grid:  board,
		conns: make(map[*websocket.Conn]struct{}),
	}
}

// Run publishes immediately, then again every second, until ctx is cancelled.
func (h *hub) Run(ctx context.Context) {
	h.publish()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			h.publish()
		}
	}
}

// frame is the shape the socket encodes. The grid returns these values.
type frame interface {
	Width() int
	Height() int
	Cells() []int
}

func marshalFrame(next frame) ([]byte, error) {
	return json.Marshal(wire{
		Width:  next.Width(),
		Height: next.Height(),
		Cells:  next.Cells(),
	})
}

func (h *hub) frame() []byte {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.last != nil {
		return h.last
	}
	payload, err := marshalFrame(h.grid.Next())
	if err != nil {
		log.Printf("grid marshal: %v", err)
		return nil
	}
	h.last = payload
	return payload
}

func (h *hub) publish() {
	payload, err := marshalFrame(h.grid.Next())
	if err != nil {
		log.Printf("grid marshal: %v", err)
		return
	}

	h.mu.Lock()
	h.last = payload
	conns := make([]*websocket.Conn, 0, len(h.conns))
	for conn := range h.conns {
		conns = append(conns, conn)
	}
	h.mu.Unlock()

	for _, conn := range conns {
		if err := writeFrame(context.Background(), conn, payload); err != nil {
			h.drop(conn)
		}
	}
}

// ServeHTTP upgrades GET /ws and keeps the socket open until the client leaves.
func (h *hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"127.0.0.1:*", "localhost:*"},
	})
	if err != nil {
		log.Printf("websocket accept: %v", err)
		return
	}
	conn.SetReadLimit(1024)

	h.mu.Lock()
	h.conns[conn] = struct{}{}
	h.mu.Unlock()

	if err := writeFrame(r.Context(), conn, h.frame()); err != nil {
		h.drop(conn)
		return
	}

	for {
		_, _, err := conn.Read(r.Context())
		if err != nil {
			h.drop(conn)
			return
		}
	}
}

func (h *hub) drop(conn *websocket.Conn) {
	h.mu.Lock()
	delete(h.conns, conn)
	h.mu.Unlock()
	conn.Close(websocket.StatusGoingAway, "")
}

func writeFrame(ctx context.Context, conn *websocket.Conn, payload []byte) error {
	writeCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	return conn.Write(writeCtx, websocket.MessageText, payload)
}
