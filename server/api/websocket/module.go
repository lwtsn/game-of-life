package websocket

import "go.uber.org/fx"

// Module provides Handler. It needs a grid and a users.Tracker.
var Module = fx.Module("websocket",
	fx.Provide(New),
)
