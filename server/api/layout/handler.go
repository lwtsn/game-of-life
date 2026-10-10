package layout

import (
	"context"

	"connectrpc.com/connect/v2"
	"game_of_life/server/api/websocket"
	lifepb "game_of_life/server/gen/life/v1"
	"game_of_life/server/internal/grid"
	"game_of_life/server/internal/layout"
	"game_of_life/server/internal/user"
)

type handler struct {
	board   grid.Grid
	people  user.Service
	sockets websocket.Handler
	shapes  layout.Catalogue
	origin  layout.Origin
}

// ListPatterns returns the catalogue Place stamps from. It needs no session: the shapes are the same for everyone.
func (h *handler) ListPatterns(context.Context, *lifepb.ListPatternsRequest) (*lifepb.ListPatternsResponse, error) {
	shapes := h.shapes.Shapes()
	out := make([]*lifepb.Shape, len(shapes))
	for i, shape := range shapes {
		cells := make([]*lifepb.Offset, len(shape.Cells))
		for j, cell := range shape.Cells {
			cells[j] = &lifepb.Offset{X: int32(cell.X), Y: int32(cell.Y)}
		}
		out[i] = &lifepb.Shape{Pattern: shape.Pattern, Label: shape.Label, Cells: cells}
	}
	return &lifepb.ListPatternsResponse{Shapes: out}, nil
}

func (h *handler) Place(ctx context.Context, req *lifepb.PlaceRequest) (*lifepb.PlaceResponse, error) {
	if req == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, "pattern is required")
	}
	shape, ok := h.shapes.ByPattern(req.GetPattern())
	if !ok {
		return nil, connect.NewError(connect.CodeInvalidArgument, "pattern is required")
	}
	id, ok := sessionID(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeInvalidArgument, "session is required")
	}
	person, ok := h.people.ByID(id)
	if !ok {
		return nil, connect.NewError(connect.CodeInvalidArgument, "session is required")
	}
	frame := h.board.Current()
	if frame == nil {
		return nil, connect.NewError(connect.CodeInternal, "board is not ready")
	}
	if shape.Width > frame.Width() || shape.Height > frame.Height() || shape.Width < 1 || shape.Height < 1 {
		return nil, connect.NewError(connect.CodeInvalidArgument, "pattern does not fit the board")
	}
	x, y := h.origin(frame.Width(), frame.Height(), shape.Width, shape.Height)
	if req.GetOrigin() != nil {
		x = int(req.GetOrigin().GetX())
		y = int(req.GetOrigin().GetY())
	}
	points := make([]grid.Point, len(shape.Cells))
	for i, cell := range shape.Cells {
		points[i] = grid.Point{X: x + cell.X, Y: y + cell.Y}
	}
	if !h.board.PlaceAll(points, person) {
		return nil, connect.NewError(connect.CodeInvalidArgument, "pattern does not fit the board")
	}
	h.sockets.Publish()
	return &lifepb.PlaceResponse{
		Pattern: shape.Pattern,
		Origin:  &lifepb.Offset{X: int32(x), Y: int32(y)},
		Applied: true,
	}, nil
}

func sessionID(ctx context.Context) (string, bool) {
	info, ok := connect.CallInfoForServerContext(ctx)
	if !ok {
		return "", false
	}
	id := info.RequestHeader().Get("X-Session")
	if id == "" {
		return "", false
	}
	return id, true
}
