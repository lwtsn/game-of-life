package grid

import (
	"math/rand/v2"
	"time"
)

const (
	Cols = 80
	Rows = 50
)

// Snapshot is the provisional frame sent on the websocket once a second.
// width, height, and cells are the whole contract until we settle it.
type Snapshot struct {
	Width  int   `json:"width"`
	Height int   `json:"height"`
	Cells  []int `json:"cells"`
}

// Source produces the next full grid.
type Source interface {
	Next() Snapshot
}

// Random fills the whole 80 by 50 grid. About one cell in six is alive,
// so a frame is readable instead of solid noise.
type Random struct {
	rng *rand.Rand
}

func NewRandom() *Random {
	return NewRandomSeed(time.Now().UnixNano())
}

func NewRandomSeed(seed int64) *Random {
	return &Random{rng: rand.New(rand.NewPCG(uint64(seed), uint64(seed>>1|1)))}
}

func (r *Random) Next() Snapshot {
	cells := make([]int, Cols*Rows)
	for i := range cells {
		if r.rng.IntN(6) == 0 {
			cells[i] = 1
		}
	}
	return Snapshot{Width: Cols, Height: Rows, Cells: cells}
}
