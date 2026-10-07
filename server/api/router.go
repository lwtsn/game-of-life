package api

import "github.com/gin-gonic/gin"

// Register mounts the downstream handlers.
func (h *handler) Register(r *gin.Engine) {
	r.GET("/ws", h.sockets.Serve)
	r.POST("/start", h.controls.Start)
	r.POST("/layout", h.controls.Layout)
}
