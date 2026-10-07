package api

import (
	"context"

	"github.com/gin-gonic/gin"
)

// Handler registers the HTTP routes and listens for board updates.
type Handler interface {
	Register(*gin.Engine)
	Run(context.Context)
}
