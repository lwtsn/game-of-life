package api

import (
	"game_of_life/server/api/controls"
	"game_of_life/server/api/users"
	"game_of_life/server/api/websocket"

	"go.uber.org/fx"
)

// Module wires the route packages into Fx. Callers receive a Handler.
var Module = fx.Module("api",
	users.Module,
	websocket.Module,
	controls.Module,
	fx.Provide(provideHandler),
)

func provideHandler(sockets websocket.Handler, ctrl controls.Handler) Handler {
	return &handler{sockets: sockets, controls: ctrl}
}
