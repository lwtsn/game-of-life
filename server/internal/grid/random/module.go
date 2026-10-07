package random

import (
	"game_of_life/server/internal/grid/source"
	"go.uber.org/fx"
)

// Module wires this random filler into Fx. Callers receive a source.Source.
var Module = fx.Module("random",
	fx.Provide(provide),
)

func provide() source.Source {
	return newRandom()
}
