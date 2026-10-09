package websocket

import (
	"context"
	"log"
	"time"

	lifepb "game_of_life/server/gen/life/v1"

	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/encoding/protojson"
)

func (h *handler) Serve(c *gin.Context) {
	conn, err := websocket.Accept(c.Writer, c.Request, &websocket.AcceptOptions{
		OriginPatterns: h.origins.Patterns(),
	})
	if err != nil {
		log.Printf("websocket accept: %v", err)
		return
	}
	conn.SetReadLimit(1024)

	fresh := h.track(c.Request, conn)

	payload, err := h.grid.Current().ToJson()
	if err != nil {
		log.Printf("grid json: %v", err)
		h.disconnect(conn)
		return
	}
	if err := h.send(c.Request.Context(), conn, payload); err != nil {
		h.disconnect(conn)
		return
	}
	h.broadcastPeople()
	if person := h.userFor(conn); person != nil {
		h.tellColour(person.ID(), person.Colour())
		if fresh {
			h.announce(lifepb.MessageType_MESSAGE_TYPE_ENTERED, person.Colour())
		}
	}
	h.broadcastClock()

	for {
		_, data, err := conn.Read(c.Request.Context())
		if err != nil {
			h.disconnect(conn)
			h.broadcastPeople()
			return
		}
		h.handle(conn, data)
	}
}

func (h *handler) handle(conn *websocket.Conn, data []byte) {
	var msg lifepb.ClientMessage
	if err := protojson.Unmarshal(data, &msg); err != nil {
		return
	}
	switch action := msg.GetAction().(type) {
	case *lifepb.ClientMessage_ResetBoard:
		if action.ResetBoard {
			h.reset(conn)
		}
		return
	case *lifepb.ClientMessage_Colour:
		if action.Colour != "" {
			h.chooseColour(conn, action.Colour)
		}
		return
	case *lifepb.ClientMessage_Pace:
		h.setPace(conn, action.Pace)
		return
	case *lifepb.ClientMessage_Playback:
		h.setRunning(conn, action.Playback)
		return
	case *lifepb.ClientMessage_Point:
		if action.Point == nil {
			return
		}
		person := h.userFor(conn)
		if person == nil {
			return
		}
		if !h.grid.Place(int(action.Point.GetX()), int(action.Point.GetY()), person) {
			return
		}
	default:
		return
	}
	payload, err := h.grid.Current().ToJson()
	if err != nil {
		log.Printf("grid json: %v", err)
		return
	}
	h.writeAll(payload)
}

func (h *handler) reset(conn *websocket.Conn) {
	person := h.userFor(conn)
	if person == nil {
		return
	}
	h.grid.SetRunning(false)
	if !h.grid.Clear() {
		h.broadcastClock()
		return
	}
	payload, err := h.grid.Current().ToJson()
	if err != nil {
		log.Printf("grid json: %v", err)
		return
	}
	h.writeAll(payload)
	h.announce(lifepb.MessageType_MESSAGE_TYPE_RESET, person.Colour())
	h.broadcastClock()
}

func (h *handler) setPace(conn *websocket.Conn, pace int32) {
	if h.userFor(conn) == nil || !knownPace(pace) {
		return
	}
	if !h.grid.SetPace(int(pace)) {
		return
	}
	h.broadcastClock()
}

func (h *handler) setRunning(conn *websocket.Conn, playback lifepb.Playback) {
	if h.userFor(conn) == nil {
		return
	}
	switch playback {
	case lifepb.Playback_PLAYBACK_RUNNING:
		h.grid.SetRunning(true)
	case lifepb.Playback_PLAYBACK_STOPPED:
		h.grid.SetRunning(false)
	default:
		return
	}
	h.broadcastClock()
}

func knownPace(pace int32) bool {
	return pace >= int32(lifepb.PaceBound_PACE_BOUND_MIN) && pace <= int32(lifepb.PaceBound_PACE_BOUND_MAX)
}

func (h *handler) broadcastClock() {
	running, pace := h.grid.Clock()
	playback := lifepb.Playback_PLAYBACK_STOPPED
	if running {
		playback = lifepb.Playback_PLAYBACK_RUNNING
	}
	payload, err := protojson.Marshal(&lifepb.ServerMessage{
		Type:     lifepb.MessageType_MESSAGE_TYPE_CLOCK,
		Pace:     int32(pace),
		Playback: playback,
	})
	if err != nil {
		log.Printf("clock json: %v", err)
		return
	}
	h.writeAll(payload)
}

func (h *handler) chooseColour(conn *websocket.Conn, colour string) {
	person := h.userFor(conn)
	if person == nil {
		return
	}
	next, ok := h.svc.SetColour(person.ID(), colour)
	if !ok {
		return
	}
	h.tellColour(next.ID(), next.Colour())
	h.broadcastPeople()
	h.announce(lifepb.MessageType_MESSAGE_TYPE_CHANGED, next.Colour())
}

func (h *handler) announce(kind lifepb.MessageType, colour string) {
	if colour == "" {
		return
	}
	payload, err := protojson.Marshal(&lifepb.ServerMessage{Type: kind, Colour: colour})
	if err != nil {
		log.Printf("presence json: %v", err)
		return
	}
	h.writeAll(payload)
}

func (h *handler) tellColour(id, colour string) {
	payload, err := protojson.Marshal(&lifepb.ServerMessage{
		Type:   lifepb.MessageType_MESSAGE_TYPE_YOU,
		Colour: colour,
	})
	if err != nil {
		log.Printf("colour json: %v", err)
		return
	}
	for _, conn := range h.conns() {
		person := h.stored(conn)
		if person == nil || person.ID() != id {
			continue
		}
		if err := h.send(context.Background(), conn, payload); err != nil {
			h.disconnect(conn)
		}
	}
}

func writeFrame(ctx context.Context, conn *websocket.Conn, payload []byte) error {
	writeCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	return conn.Write(writeCtx, websocket.MessageText, payload)
}
