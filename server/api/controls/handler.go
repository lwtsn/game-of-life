package controls

import (
	"context"
	"sync"

	"game_of_life/server/internal/grid"

	"github.com/gin-gonic/gin"
)

// Handler is the simulation control routes.
type Handler interface {
	Start(*gin.Context)
	Layout(*gin.Context)
	Bind(context.Context)
}

type handler struct {
	grid grid.Grid

	mu  sync.Mutex
	ctx context.Context
}

func New(board grid.Grid) Handler {
	return &handler{grid: board}
}

// Bind keeps the process context so Start can cancel with the server.
func (h *handler) Bind(ctx context.Context) {
	h.mu.Lock()
	h.ctx = ctx
	h.mu.Unlock()
}
