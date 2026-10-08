package websocket

import "go.uber.org/fx"

// Module provides Handler. It needs a grid and a user service.
var Module = fx.Module("websocket",
	fx.Provide(New),
)
