package life

import (
	"game_of_life/server/internal/grid/source"

	"go.uber.org/fx"
)

// Module provides a source.Source.
var Module = fx.Module("life",
	fx.Provide(provide),
)

func provide() source.Source {
	return newLife()
}
