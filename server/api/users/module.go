package users

import "go.uber.org/fx"

// Module provides Tracker.
var Module = fx.Module("users",
	fx.Provide(New),
)
