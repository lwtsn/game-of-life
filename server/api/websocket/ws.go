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

	h.track(c.Request, conn)

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
		X *int `json:"x"`
		Y *int `json:"y"`
	}
	if err := json.Unmarshal(data, &body); err != nil {
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

func writeFrame(ctx context.Context, conn *websocket.Conn, payload []byte) error {
	writeCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	return conn.Write(writeCtx, websocket.MessageText, payload)
}
