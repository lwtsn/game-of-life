package grid

import "go.uber.org/fx"

// Module wires the random source into Fx. Callers receive a Source.
var Module = fx.Module("grid",
	fx.Provide(provideSource),
)

func provideSource() Source {
	return newRandom()
}
