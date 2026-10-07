package grid

import (
	"game_of_life/server/internal/grid/source"
	"go.uber.org/fx"
)

// Module wires a grid that reads whatever source.Source Fx supplies.
var Module = fx.Module("grid",
	fx.Provide(provide),
)

func provide(src source.Source) Grid {
	return newGrid(src)
}
