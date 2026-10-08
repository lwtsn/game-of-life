package api

import (
	"context"

	"game_of_life/server/api/websocket"

	"github.com/gin-gonic/gin"
)

// Handler registers the HTTP routes and listens for board updates.
type Handler interface {
	Register(*gin.Engine)
	Run(context.Context)
}

type handler struct {
	sockets websocket.Handler
}

func (h *handler) Run(ctx context.Context) {
	h.sockets.Run(ctx)
}
