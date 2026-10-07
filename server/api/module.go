package api

import (
	"game_of_life/server/internal/grid"
	"go.uber.org/fx"
)

// Module wires the hub into Fx. Callers receive a Handler.
var Module = fx.Module("api",
	fx.Provide(provideHandler),
)

func provideHandler(source grid.Source) Handler {
	return newHub(source)
}
