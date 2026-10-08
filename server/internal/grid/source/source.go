package source

import (
	lifepb "game_of_life/server/gen/life/v1"
	"game_of_life/server/internal/user"

	"google.golang.org/protobuf/encoding/protojson"
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
	return protojson.Marshal(cellMessage(c))
}

func cellMessage(c Cell) *lifepb.Cell {
	colour := c.Colour
	alive := c.Alive
	id := ""
	if c.User != nil {
		alive = true
		id = c.User.ID()
		if colour == "" {
			colour = c.User.Colour()
		}
	}
	return &lifepb.Cell{Alive: alive, Id: id, Colour: colour}
}

// Frame is one full grid.
type Frame interface {
	Width() int
	Height() int
	Cells() []Cell
	ToJson() ([]byte, error)
}

// Encode is the board message: type, width, height, cells, and the generation.
// generation is the clock's frame. Zero is the board before the first step.
func Encode(frame Frame, generation int) ([]byte, error) {
	cells := frame.Cells()
	body := make([]*lifepb.Cell, len(cells))
	for i, cell := range cells {
		body[i] = cellMessage(cell)
	}
	return protojson.Marshal(&lifepb.ServerMessage{
		Type:   lifepb.MessageType_MESSAGE_TYPE_BOARD,
		Width:  int32(frame.Width()),
		Height: int32(frame.Height()),
		Cells:  body,
		Frame:  int32(generation),
	})
}

// Source produces the next values from the board the grid stores.
// current is that board. It is nil on the first step.
type Source interface {
	Next(current Frame) Frame
}
