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
