package websocket

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
)

func (h *handler) Serve(c *gin.Context) {
	conn, err := websocket.Accept(c.Writer, c.Request, &websocket.AcceptOptions{
		OriginPatterns: []string{"127.0.0.1:*", "localhost:*"},
	})
	if err != nil {
		log.Printf("websocket accept: %v", err)
		return
	}
	conn.SetReadLimit(1024)

	fresh := h.track(c.Request, conn)

	payload, err := h.grid.Current().ToJson()
	if err != nil {
		log.Printf("grid json: %v", err)
		h.disconnect(conn)
		return
	}
	if err := h.send(c.Request.Context(), conn, payload); err != nil {
		h.disconnect(conn)
		return
	}
	h.broadcastColours()
	if person := h.userFor(conn); person != nil {
		h.tellColour(person.ID(), person.Colour())
		if fresh {
			h.announce("entered", person.Colour())
		}
	}

	for {
		_, data, err := conn.Read(c.Request.Context())
		if err != nil {
			h.disconnect(conn)
			h.broadcastColours()
			return
		}
		h.handle(conn, data)
	}
}

func (h *handler) handle(conn *websocket.Conn, data []byte) {
	var body struct {
		X      *int   `json:"x"`
		Y      *int   `json:"y"`
		Reset  bool   `json:"reset"`
		Colour string `json:"colour"`
	}
	if err := json.Unmarshal(data, &body); err != nil {
		return
	}
	if body.Reset {
		h.reset(conn)
		return
	}
	if body.Colour != "" {
		h.chooseColour(conn, body.Colour)
		return
	}
	if body.X == nil || body.Y == nil {
		return
	}
	person := h.userFor(conn)
	if person == nil {
		return
	}
	if !h.grid.Place(*body.X, *body.Y, person) {
		return
	}
	payload, err := h.grid.Current().ToJson()
	if err != nil {
		log.Printf("grid json: %v", err)
		return
	}
	h.writeAll(payload)
}

func (h *handler) reset(conn *websocket.Conn) {
	person := h.userFor(conn)
	if person == nil {
		return
	}
	if !h.grid.Clear() {
		return
	}
	payload, err := h.grid.Current().ToJson()
	if err != nil {
		log.Printf("grid json: %v", err)
		return
	}
	h.writeAll(payload)
	notice, err := json.Marshal(struct {
		Reset string `json:"reset"`
	}{Reset: person.Colour()})
	if err != nil {
		log.Printf("reset json: %v", err)
		return
	}
	h.writeAll(notice)
}

func (h *handler) chooseColour(conn *websocket.Conn, colour string) {
	person := h.userFor(conn)
	if person == nil {
		return
	}
	next, ok := h.svc.SetColour(person.ID(), colour)
	if !ok {
		return
	}
	h.tellColour(next.ID(), next.Colour())
	h.broadcastColours()
}

func (h *handler) announce(field, colour string) {
	if colour == "" {
		return
	}
	payload, err := json.Marshal(map[string]string{field: colour})
	if err != nil {
		log.Printf("presence json: %v", err)
		return
	}
	h.writeAll(payload)
}

func (h *handler) tellColour(id, colour string) {
	payload, err := json.Marshal(struct {
		You string `json:"you"`
	}{You: colour})
	if err != nil {
		log.Printf("colour json: %v", err)
		return
	}
	for _, conn := range h.conns() {
		person := h.stored(conn)
		if person == nil || person.ID() != id {
			continue
		}
		if err := h.send(context.Background(), conn, payload); err != nil {
			h.disconnect(conn)
		}
	}
}

func writeFrame(ctx context.Context, conn *websocket.Conn, payload []byte) error {
	writeCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	return conn.Write(writeCtx, websocket.MessageText, payload)
}
