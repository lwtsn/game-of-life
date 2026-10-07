package api

import (
	"game_of_life/server/internal/grid"
	"go.uber.org/fx"
)

// Module provides a Handler. The hub stays in this package.
var Module = fx.Module("api",
	fx.Provide(provideHandler),
)

func provideHandler(source grid.Source) Handler {
	return newHub(source)
}
