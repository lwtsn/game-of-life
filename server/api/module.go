package api

import (
	"game_of_life/server/internal/grid/source"
	"go.uber.org/fx"
)

// Module wires the hub into Fx. Callers receive a Handler.
var Module = fx.Module("api",
	fx.Provide(provideHandler),
)

func provideHandler(source source.Source) Handler {
	return newHub(source)
}
