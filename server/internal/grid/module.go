package grid

import "go.uber.org/fx"

// Module provides a Source. The implementation stays in this package.
var Module = fx.Module("grid",
	fx.Provide(provideSource),
)

func provideSource() Source {
	return newRandom()
}
