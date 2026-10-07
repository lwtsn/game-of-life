package controls

import "go.uber.org/fx"

// Module provides Handler.
var Module = fx.Module("controls",
	fx.Provide(New),
)
