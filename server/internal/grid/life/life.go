package life

import "game_of_life/server/internal/grid/source"

const (
	cols = 80
	rows = 50
)

// life is a source.Source. It returns the next board from the one it is given.
type life struct{}

func newLife() source.Source {
	return life{}
}

type snapshot struct {
	width  int
	height int
	cells  []source.Cell
}

func (s snapshot) Width() int           { return s.width }
func (s snapshot) Height() int          { return s.height }
func (s snapshot) Cells() []source.Cell { return s.cells }
func (s snapshot) ToJson() ([]byte, error) {
	return source.Encode(s)
}

func (life) Next(current source.Frame) source.Frame {
	if current == nil {
		return snapshot{width: cols, height: rows, cells: make([]source.Cell, cols*rows)}
	}
	width := current.Width()
	height := current.Height()
	cells := current.Cells()
	next := make([]source.Cell, width*height)
	for y := range height {
		for x := range width {
			next[y*width+x] = nextCell(cells, width, height, x, y)
		}
	}
	return snapshot{width: width, height: height, cells: next}
}

// nextCell applies the four rules to one square. A neighbour is one of the eight
// squares that sit inside the board. A cell that survives keeps the colour on
// it and, when a person placed it, that person. A birth is alive, has no
// person, and takes the average of the three colours around it.
func nextCell(cells []source.Cell, width, height, x, y int) source.Cell {
	n := liveNeighbours(cells, width, height, x, y)
	current := cells[y*width+x]
	if n == 3 || (current.Alive && n == 2) {
		if current.Alive {
			return kept(current)
		}
		return born(cells, width, height, x, y)
	}
	return source.Cell{}
}

func kept(current source.Cell) source.Cell {
	if current.Colour == "" && current.User != nil {
		current.Colour = current.User.Colour()
	}
	return current
}

func born(cells []source.Cell, width, height, x, y int) source.Cell {
	colours := make([]string, 0, 3)
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			nx, ny := x+dx, y+dy
			if nx < 0 || ny < 0 || nx >= width || ny >= height {
				continue
			}
			parent := cells[ny*width+nx]
			if !parent.Alive {
				continue
			}
			colour := parent.Colour
			if colour == "" && parent.User != nil {
				colour = parent.User.Colour()
			}
			if colour == "" {
				return source.Cell{Alive: true}
			}
			colours = append(colours, colour)
		}
	}
	if len(colours) != 3 {
		return source.Cell{Alive: true}
	}
	mixed, err := BirthColour(colours[0], colours[1], colours[2])
	if err != nil {
		return source.Cell{Alive: true}
	}
	return source.Cell{Alive: true, Colour: mixed}
}

func liveNeighbours(cells []source.Cell, width, height, x, y int) int {
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

func cellAt(cells []source.Cell, width, height, x, y int) int {
	if x < 0 || y < 0 || x >= width || y >= height {
		return 0
	}
	if cells[y*width+x].Alive {
		return 1
	}
	return 0
}
