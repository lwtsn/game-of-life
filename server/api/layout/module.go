package layout

import (
	"connectrpc.com/connect/v2"
	"game_of_life/server/api/websocket"
	lifepbconnect "game_of_life/server/gen/life/v1/lifepbconnect"
	"game_of_life/server/internal/grid"
	"game_of_life/server/internal/layout"
	"game_of_life/server/internal/user"

	"go.uber.org/fx"
)

// Module provides the Connect server for LayoutService.
var Module = fx.Module("api.layout",
	layout.Module,
	fx.Provide(provideHandler, provideServer),
)

func provideHandler(
	board grid.Grid,
	people user.Service,
	sockets websocket.Handler,
	shapes layout.Catalogue,
	origin layout.Origin,
) lifepbconnect.LayoutServiceHandler {
	return &handler{
		board:   board,
		people:  people,
		sockets: sockets,
		shapes:  shapes,
		origin:  origin,
	}
}

func provideServer(svc lifepbconnect.LayoutServiceHandler) *connect.Server {
	server := connect.NewServer()
	lifepbconnect.RegisterLayoutServiceHandler(server, svc)
	return server
}
