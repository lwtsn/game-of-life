package source

// Frame is one full grid.
type Frame interface {
	Width() int
	Height() int
	Cells() []int
}

// Source produces the next values from the board the grid stores.
// current is that board. It is nil on the first step.
type Source interface {
	Next(current Frame) Frame
}
