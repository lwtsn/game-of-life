package grid

// Frame is one full grid. snapshot is the value a source returns.
type Frame interface {
	Width() int
	Height() int
	Cells() []int
}

type snapshot struct {
	width  int
	height int
	cells  []int
}

func (s snapshot) Width() int   { return s.width }
func (s snapshot) Height() int  { return s.height }
func (s snapshot) Cells() []int { return s.cells }
