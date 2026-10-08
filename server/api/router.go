package api

import "github.com/gin-gonic/gin"

// Register mounts the board socket. Patterns go through Connect Place.
func (h *handler) Register(r *gin.Engine) {
	r.GET("/ws", h.sockets.Serve)
}
