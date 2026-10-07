package random

import (
	"math/rand/v2"
	"time"

	"game_of_life/server/internal/grid/source"
)

const (
	cols = 80
	rows = 50
)

// random is a source.Source. About one cell in six is alive, so a frame is
// readable instead of solid noise.
type random struct {
	rng *rand.Rand
}

type snapshot struct {
	width  int
	height int
	cells  []int
}

func (s snapshot) Width() int   { return s.width }
func (s snapshot) Height() int  { return s.height }
func (s snapshot) Cells() []int { return s.cells }

func newRandom() *random {
	return newRandomSeed(time.Now().UnixNano())
}

func newRandomSeed(seed int64) *random {
	return &random{rng: rand.New(rand.NewPCG(uint64(seed), uint64(seed>>1|1)))}
}

func (r *random) Next(source.Frame) source.Frame {
	cells := make([]int, cols*rows)
	for i := range cells {
		if r.rng.IntN(6) == 0 {
			cells[i] = 1
		}
	}
	return snapshot{width: cols, height: rows, cells: cells}
}
