package random

import (
	"game_of_life/server/internal/grid/source"
	"go.uber.org/fx"
)

// Module provides a source.Source. The grid reads it. A life source can replace this module.
var Module = fx.Module("random",
	fx.Provide(provide),
)

func provide() source.Source {
	return newRandom()
}
