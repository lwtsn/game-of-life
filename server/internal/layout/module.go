package layout

import "go.uber.org/fx"

// Module provides Catalogue and Origin.
var Module = fx.Module("layout",
	fx.Provide(loadCatalogue, provideOrigin),
)

func provideOrigin() Origin { return randomOrigin }
