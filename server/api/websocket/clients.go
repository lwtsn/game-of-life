package websocket

import (
	"net/http"

	lifepb "game_of_life/server/gen/life/v1"
	"game_of_life/server/internal/user"

	"github.com/coder/websocket"
)

// track joins the session. The bool is true only for that session's first socket.
func (h *handler) track(r *http.Request, conn *websocket.Conn) bool {
	person, _ := h.svc.Join(r.URL.Query().Get("session"))
	h.mu.Lock()
	defer h.mu.Unlock()
	fresh := person != nil
	if fresh {
		for _, other := range h.clients {
			if other != nil && other.ID() == person.ID() {
				fresh = false
				break
			}
		}
	}
	h.clients[conn] = person
	return fresh
}

func (h *handler) stored(conn *websocket.Conn) user.User {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.clients[conn]
}

func (h *handler) userFor(conn *websocket.Conn) user.User {
	person := h.stored(conn)
	if person == nil {
		return nil
	}
	current, ok := h.svc.ByID(person.ID())
	if !ok {
		return person
	}
	return current
}

func (h *handler) conns() []*websocket.Conn {
	h.mu.Lock()
	defer h.mu.Unlock()
	conns := make([]*websocket.Conn, 0, len(h.clients))
	for conn := range h.clients {
		conns = append(conns, conn)
	}
	return conns
}

func (h *handler) disconnect(conn *websocket.Conn) {
	h.mu.Lock()
	person, found := h.clients[conn]
	if found {
		delete(h.clients, conn)
	}
	stillHere := false
	if person != nil {
		for _, other := range h.clients {
			if other != nil && other.ID() == person.ID() {
				stillHere = true
				break
			}
		}
	}
	h.mu.Unlock()
	if !found {
		return
	}
	if person != nil && !stillHere {
		colour := person.Colour()
		if current, ok := h.svc.ByID(person.ID()); ok {
			colour = current.Colour()
		}
		h.svc.Leave(person.ID())
		h.announce(lifepb.MessageType_MESSAGE_TYPE_EXITED, colour)
	}
	_ = conn.Close(websocket.StatusGoingAway, "")
}
