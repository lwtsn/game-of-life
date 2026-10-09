package websocket

import (
	"context"
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
			if other.person != nil && other.person.ID() == person.ID() {
				fresh = false
				break
			}
		}
	}
	c := &client{person: person, out: make(chan []byte, mailboxSize)}
	h.clients[conn] = c
	go h.pump(conn, c.out)
	return fresh
}

// pump is the only goroutine that writes to conn. It sends queued messages in order,
// so a slow socket only ever delays itself. The first failed write drops the socket.
func (h *handler) pump(conn *websocket.Conn, out <-chan []byte) {
	for payload := range out {
		if err := writeFrame(context.Background(), conn, payload); err != nil {
			if h.disconnect(conn) {
				h.broadcastPeople()
			}
			return
		}
	}
}

func (h *handler) stored(conn *websocket.Conn) user.User {
	h.mu.Lock()
	defer h.mu.Unlock()
	c, ok := h.clients[conn]
	if !ok {
		return nil
	}
	return c.person
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

// disconnect forgets the socket and closes it. It reports false if the socket was already gone.
func (h *handler) disconnect(conn *websocket.Conn) bool {
	h.mu.Lock()
	c, found := h.clients[conn]
	var person user.User
	if found {
		delete(h.clients, conn)
		// Closed under mu, the same lock senders hold, so nothing can send on it afterwards.
		close(c.out)
		person = c.person
	}
	stillHere := false
	if person != nil {
		for _, other := range h.clients {
			if other.person != nil && other.person.ID() == person.ID() {
				stillHere = true
				break
			}
		}
	}
	h.mu.Unlock()
	if !found {
		return false
	}
	if person != nil && !stillHere {
		colour := person.Colour()
		if current, ok := h.svc.ByID(person.ID()); ok {
			colour = current.Colour()
		}
		h.svc.Leave(person.ID())
		h.announce(lifepb.MessageType_MESSAGE_TYPE_EXITED, colour)
	}
	// The close handshake waits on the peer, which may be the slow socket, so it runs on its own.
	go func() { _ = conn.Close(websocket.StatusGoingAway, "") }()
	return true
}
