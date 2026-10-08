package user

import "go.uber.org/fx"

// Module provides Service.
var Module = fx.Module("user",
	fx.Provide(NewService),
)
