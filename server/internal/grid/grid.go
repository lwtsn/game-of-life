package grid

import (
	"context"
	"log"
	"sync"
	"time"

	lifepb "game_of_life/server/gen/life/v1"
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
	Clear() bool
	SetPace(perSecond int) bool
	SetRunning(on bool)
	Clock() (running bool, pace int)
	// Live reports whether the board waiting in Updates is still the one to send.
	Live() bool
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
	frame  int
}

func (s snapshot) Width() int           { return s.width }
func (s snapshot) Height() int          { return s.height }
func (s snapshot) Cells() []source.Cell { return s.cells }
func (s snapshot) ToJson() ([]byte, error) {
	return source.Encode(s, s.frame)
}

type grid struct {
	src     source.Source
	current source.Frame
	updates chan []byte
	wake    chan struct{}

	mu         sync.Mutex
	started    bool
	simulating bool
	perSecond  int
	frame      int
	seq        uint64
	queuedSeq  uint64
}

func newGrid(src source.Source) *grid {
	g := &grid{
		src:        src,
		updates:    make(chan []byte, 1),
		wake:       make(chan struct{}, 1),
		simulating: true,
		perSecond:  int(lifepb.PaceBound_PACE_BOUND_MIN),
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
	g.take(g.src.Next(g.current))
}

func (g *grid) take(next source.Frame) {
	g.current = snapshot{
		width:  next.Width(),
		height: next.Height(),
		cells:  append([]source.Cell(nil), next.Cells()...),
		frame:  g.frame,
	}
}

func (g *grid) Updates() <-chan []byte {
	return g.updates
}

// placed records the person and copies their colour onto the square.
func placed(person user.User) source.Cell {
	return source.Cell{Alive: true, User: person, Colour: person.Colour()}
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
	cells[y*width+x] = placed(person)
	g.current = snapshot{width: width, height: height, cells: cells, frame: g.frame}
	g.supersedeLocked()
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
		cells[point.Y*width+point.X] = placed(person)
	}
	g.current = snapshot{width: width, height: height, cells: cells, frame: g.frame}
	g.supersedeLocked()
	return true
}

// Clear kills every cell and keeps the width and height.
// It returns false when there is no board.
func (g *grid) Clear() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.current == nil {
		return false
	}
	width := g.current.Width()
	height := g.current.Height()
	g.current = snapshot{
		width:  width,
		height: height,
		cells:  make([]source.Cell, width*height),
		frame:  g.frame,
	}
	g.supersedeLocked()
	return true
}

// SetPace stores how many generations pass in one second.
// It returns false when perSecond is outside the shared bounds.
func (g *grid) SetPace(perSecond int) bool {
	if perSecond < int(lifepb.PaceBound_PACE_BOUND_MIN) || perSecond > int(lifepb.PaceBound_PACE_BOUND_MAX) {
		return false
	}
	g.mu.Lock()
	g.perSecond = perSecond
	g.mu.Unlock()
	g.ping()
	return true
}

// SetRunning starts or stops the clock. The loop keeps running either way.
// Stopping discards a board the hub has not taken yet.
func (g *grid) SetRunning(on bool) {
	g.mu.Lock()
	g.simulating = on
	if !on {
		g.discardLocked()
	}
	g.mu.Unlock()
	g.ping()
}

func (g *grid) discardLocked() {
	for {
		select {
		case <-g.updates:
		default:
			return
		}
	}
}

// supersedeLocked drops a generation queued before this edit, so the hub does not paint it afterwards.
func (g *grid) supersedeLocked() {
	g.seq++
	g.discardLocked()
}

// Live reports whether the payload just taken from Updates is still current.
func (g *grid) Live() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.queuedSeq == g.seq
}

// Clock reports whether the board is stepping and the generations per second.
func (g *grid) Clock() (running bool, pace int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.simulating, g.perSecond
}

func (g *grid) ping() {
	select {
	case g.wake <- struct{}{}:
	default:
	}
}

func (g *grid) interval() time.Duration {
	g.mu.Lock()
	defer g.mu.Unlock()
	return time.Second / time.Duration(g.perSecond)
}

func (g *grid) simulatingNow() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.simulating
}

// Start runs the simulation until ctx is cancelled. A second call does nothing.
func (g *grid) Start(ctx context.Context) {
	g.mu.Lock()
	if g.started {
		g.mu.Unlock()
		return
	}
	g.started = true
	g.mu.Unlock()

	go g.loop(ctx)
}

func (g *grid) loop(ctx context.Context) {
	g.notify()
	var ticker *time.Ticker
	var tick <-chan time.Time
	if g.simulatingNow() {
		ticker = time.NewTicker(g.interval())
		tick = ticker.C
	}
	defer func() {
		if ticker != nil {
			ticker.Stop()
		}
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case <-g.wake:
			if ticker != nil {
				ticker.Stop()
				ticker = nil
			}
			tick = nil
			if g.simulatingNow() {
				ticker = time.NewTicker(g.interval())
				tick = ticker.C
			}
		case <-tick:
			g.step()
		}
	}
}

// step advances one generation and queues it. It does nothing once the clock is stopped, and it holds the lock across the queue send so a stop cannot leave that board behind.
func (g *grid) step() {
	g.mu.Lock()
	defer g.mu.Unlock()
	if !g.simulating || g.current == nil {
		return
	}
	g.frame++
	g.take(g.src.Next(g.current))
	if len(g.updates) > 0 {
		return
	}
	payload, err := g.current.ToJson()
	if err != nil {
		log.Printf("grid json: %v", err)
		return
	}
	select {
	case g.updates <- payload:
		g.queuedSeq = g.seq
	default:
	}
}

func (g *grid) notify() {
	g.mu.Lock()
	defer g.mu.Unlock()
	if len(g.updates) > 0 || g.current == nil {
		return
	}
	payload, err := g.current.ToJson()
	if err != nil {
		log.Printf("grid json: %v", err)
		return
	}
	select {
	case g.updates <- payload:
		g.queuedSeq = g.seq
	default:
	}
}
