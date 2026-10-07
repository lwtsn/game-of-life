package source

import "encoding/json"

// Frame is one full grid.
type Frame interface {
	Width() int
	Height() int
	Cells() []int
	ToJson() ([]byte, error)
}

// Encode is the JSON payload for a frame: width, height, and cells.
func Encode(frame Frame) ([]byte, error) {
	return json.Marshal(struct {
		Width  int   `json:"width"`
		Height int   `json:"height"`
		Cells  []int `json:"cells"`
	}{
		Width:  frame.Width(),
		Height: frame.Height(),
		Cells:  frame.Cells(),
	})
}

// Source produces the next values from the board the grid stores.
// current is that board. It is nil on the first step.
type Source interface {
	Next(current Frame) Frame
}
