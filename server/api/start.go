package api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *hub) start(c *gin.Context) {
	ctx := context.Background()
	h.mu.Lock()
	if h.ctx != nil {
		ctx = h.ctx
	}
	h.mu.Unlock()

	h.grid.Start(ctx)
	c.Status(http.StatusNoContent)
}
