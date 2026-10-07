package source

// Frame is one full grid.
type Frame interface {
	Width() int
	Height() int
	Cells() []int
}

// Source produces the next full grid.
type Source interface {
	Next() Frame
}
