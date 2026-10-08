package grid

import (
	"context"
	"log"
	"sync"
	"time"

	"game_of_life/server/internal/grid/source"
	"game_of_life/server/internal/user"
)

type Grid interface {
	Current() source.Frame
	Advance()
	Start(context.Context)
	Updates() <-chan []byte
	Place(x, y int, person user.User) bool
	PlaceAll(points []Point, person user.User) bool
}

// Point is a column and a row on the board.
type Point struct {
	X int
	Y int
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

type grid struct {
	src     source.Source
	current source.Frame
	updates chan []byte

	mu      sync.Mutex
	running bool
}

func newGrid(src source.Source) *grid {
	g := &grid{
		src:     src,
		updates: make(chan []byte, 1),
	}
	g.Advance()
	return g
}

func (g *grid) Current() source.Frame {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.current
}

func (g *grid) Advance() {
	g.mu.Lock()
	defer g.mu.Unlock()
	next := g.src.Next(g.current)
	g.current = snapshot{
		width:  next.Width(),
		height: next.Height(),
		cells:  append([]source.Cell(nil), next.Cells()...),
	}
}

func (g *grid) Updates() <-chan []byte {
	return g.updates
}

// Place records that this person made the square at column x and row y alive.
// It returns false when there is no board, the person is nil, or the square is outside the board.
func (g *grid) Place(x, y int, person user.User) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.current == nil || person == nil {
		return false
	}
	width := g.current.Width()
	height := g.current.Height()
	if x < 0 || y < 0 || x >= width || y >= height {
		return false
	}
	cells := append([]source.Cell(nil), g.current.Cells()...)
	cells[y*width+x] = source.Cell{Alive: true, User: person}
	g.current = snapshot{width: width, height: height, cells: cells}
	return true
}

// PlaceAll records that this person made every point alive, in one update.
// It returns false when there is no board, the person is nil, the list is empty, or any point is outside the board.
func (g *grid) PlaceAll(points []Point, person user.User) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.current == nil || person == nil || len(points) == 0 {
		return false
	}
	width := g.current.Width()
	height := g.current.Height()
	for _, point := range points {
		if point.X < 0 || point.Y < 0 || point.X >= width || point.Y >= height {
			return false
		}
	}
	cells := append([]source.Cell(nil), g.current.Cells()...)
	for _, point := range points {
		cells[point.Y*width+point.X] = source.Cell{Alive: true, User: person}
	}
	g.current = snapshot{width: width, height: height, cells: cells}
	return true
}

// Start runs the simulation until ctx is cancelled. A second call does nothing.
func (g *grid) Start(ctx context.Context) {
	g.mu.Lock()
	if g.running {
		g.mu.Unlock()
		return
	}
	g.running = true
	g.mu.Unlock()

	go func() {
		g.notify()
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				g.Advance()
				g.notify()
			}
		}
	}()
}

func (g *grid) notify() {
	g.mu.Lock()
	current := g.current
	g.mu.Unlock()
	if current == nil {
		return
	}
	payload, err := current.ToJson()
	if err != nil {
		log.Printf("grid json: %v", err)
		return
	}
	select {
	case g.updates <- payload:
	default:
	}
}
