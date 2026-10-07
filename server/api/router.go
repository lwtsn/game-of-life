package api

import "github.com/gin-gonic/gin"

func (h *hub) Register(r *gin.Engine) {
	r.GET("/ws", h.ws)
	r.POST("/start", h.start)
	r.POST("/layout", h.layout)
}
