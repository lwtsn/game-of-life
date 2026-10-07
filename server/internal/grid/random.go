package grid

import (
	"math/rand/v2"
	"time"
)

const (
	cols = 80
	rows = 50
)

// Source produces the next full grid. random is the source this package provides.
type Source interface {
	Next() Frame
}

// About one cell in six is alive, so a frame is readable instead of solid noise.
type random struct {
	rng *rand.Rand
}

func newRandom() *random {
	return newRandomSeed(time.Now().UnixNano())
}

func newRandomSeed(seed int64) *random {
	return &random{rng: rand.New(rand.NewPCG(uint64(seed), uint64(seed>>1|1)))}
}

func (r *random) Next() Frame {
	cells := make([]int, cols*rows)
	for i := range cells {
		if r.rng.IntN(6) == 0 {
			cells[i] = 1
		}
	}
	return snapshot{width: cols, height: rows, cells: cells}
}
