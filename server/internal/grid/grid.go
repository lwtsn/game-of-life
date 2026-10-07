package grid

import "game_of_life/server/internal/grid/source"

// Grid is the board the API reads. It asks a source for the next values.
type Grid interface {
	Next() source.Frame
}

type grid struct {
	src source.Source
}

func newGrid(src source.Source) *grid {
	return &grid{src: src}
}

func (g *grid) Next() source.Frame {
	return g.src.Next()
}
