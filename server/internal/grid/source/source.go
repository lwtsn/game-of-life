package source

import (
	"encoding/json"

	"game_of_life/server/internal/user"
)

// Cell is one square. Alive is the Conway state. Colour is stored on the
// square itself. User is set when a person placed the square, and left unset
// when the rules created it.
type Cell struct {
	Alive  bool
	User   user.User
	Colour string
}

func (c Cell) MarshalJSON() ([]byte, error) {
	colour := c.Colour
	body := struct {
		Alive  bool   `json:"alive"`
		ID     string `json:"id,omitempty"`
		Colour string `json:"colour,omitempty"`
	}{Alive: c.Alive}
	if c.User != nil {
		body.Alive = true
		body.ID = c.User.ID()
		if colour == "" {
			colour = c.User.Colour()
		}
	}
	body.Colour = colour
	return json.Marshal(body)
}

// Frame is one full grid.
type Frame interface {
	Width() int
	Height() int
	Cells() []Cell
	ToJson() ([]byte, error)
}

// Encode is the JSON payload for a frame: width, height, and cells.
func Encode(frame Frame) ([]byte, error) {
	return json.Marshal(struct {
		Width  int    `json:"width"`
		Height int    `json:"height"`
		Cells  []Cell `json:"cells"`
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
