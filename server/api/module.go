package api

import (
	"game_of_life/server/api/layout"
	"game_of_life/server/api/websocket"

	"go.uber.org/fx"
)

// Module wires the route packages into Fx. Callers receive a Handler.
var Module = fx.Module("api",
	websocket.Module,
	layout.Module,
	fx.Provide(provideHandler),
)

func provideHandler(sockets websocket.Handler) Handler {
	return &handler{sockets: sockets}
}
