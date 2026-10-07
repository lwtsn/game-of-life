package api

import (
	"context"
	"log"
	"time"

	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
)

func (h *hub) ws(c *gin.Context) {
	conn, err := websocket.Accept(c.Writer, c.Request, &websocket.AcceptOptions{
		OriginPatterns: []string{"127.0.0.1:*", "localhost:*"},
	})
	if err != nil {
		log.Printf("websocket accept: %v", err)
		return
	}
	conn.SetReadLimit(1024)

	ip := clientIP(c.Request)
	h.track(ip, conn)

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

	for {
		_, _, err := conn.Read(c.Request.Context())
		if err != nil {
			h.disconnect(conn)
			return
		}
	}
}

func writeFrame(ctx context.Context, conn *websocket.Conn, payload []byte) error {
	writeCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	return conn.Write(writeCtx, websocket.MessageText, payload)
}
