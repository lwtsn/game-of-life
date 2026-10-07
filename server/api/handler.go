package api

import (
	"context"
	"net/http"
)

// Handler serves GET /ws and pushes a frame to every client.
type Handler interface {
	http.Handler
	Run(context.Context)
}
