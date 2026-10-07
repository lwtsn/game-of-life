package grid

import "game_of_life/server/internal/grid/source"

type Grid interface {
	Current() source.Frame
	Advance()
}

type snapshot struct {
	width  int
	height int
	cells  []int
}

func (s snapshot) Width() int   { return s.width }
func (s snapshot) Height() int  { return s.height }
func (s snapshot) Cells() []int { return s.cells }
func (s snapshot) ToJson() ([]byte, error) {
	return source.Encode(s)
}

type grid struct {
	src     source.Source
	current source.Frame
}

func newGrid(src source.Source) *grid {
	g := &grid{src: src}
	g.Advance()
	return g
}

func (g *grid) Current() source.Frame {
	return g.current
}

func (g *grid) Advance() {
	next := g.src.Next(g.current)
	g.current = snapshot{
		width:  next.Width(),
		height: next.Height(),
		cells:  append([]int(nil), next.Cells()...),
	}
}
