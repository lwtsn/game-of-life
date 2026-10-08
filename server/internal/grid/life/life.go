package life

import "game_of_life/server/internal/grid/source"

// life is a source.Source. It returns the next board from the one it is given.
type life struct{}

func newLife() source.Source {
	return life{}
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

func (life) Next(current source.Frame) source.Frame {
	width := current.Width()
	height := current.Height()
	cells := current.Cells()
	next := make([]int, width*height)
	for y := range height {
		for x := range width {
			next[y*width+x] = nextCell(cells, width, height, x, y)
		}
	}
	return snapshot{width: width, height: height, cells: next}
}

// nextCell applies the four rules to one square. A neighbour is one of the eight
// squares that sit inside the board.
func nextCell(cells []int, width, height, x, y int) int {
	n := liveNeighbours(cells, width, height, x, y)
	alive := cells[y*width+x] == 1
	if n == 3 || (alive && n == 2) {
		return 1
	}
	return 0
}

func liveNeighbours(cells []int, width, height, x, y int) int {
	n := 0
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			n += cellAt(cells, width, height, x+dx, y+dy)
		}
	}
	return n
}

func cellAt(cells []int, width, height, x, y int) int {
	if x < 0 || y < 0 || x >= width || y >= height {
		return 0
	}
	if cells[y*width+x] == 1 {
		return 1
	}
	return 0
}
