package websocket

import (
	"net"
	"net/http"

	"game_of_life/server/internal/user"

	"github.com/coder/websocket"
)

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func (h *handler) track(r *http.Request, conn *websocket.Conn) {
	person := h.svc.Join(clientIP(r))
	h.mu.Lock()
	h.clients[conn] = person
	h.mu.Unlock()
}

func (h *handler) userFor(conn *websocket.Conn) user.User {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.clients[conn]
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
	if found {
		for _, other := range h.clients {
			if other.IP() == person.IP() {
				stillHere = true
				break
			}
		}
	}
	h.mu.Unlock()
	if !found {
		return
	}
	if !stillHere {
		h.svc.Leave(person.IP())
	}
	_ = conn.Close(websocket.StatusGoingAway, "")
	h.drop(conn)
	h.reconnect(conn)
}

// drop and reconnect are stubs. Who gets dropped, and how a client reconnects, comes later.
func (h *handler) drop(*websocket.Conn)      {}
func (h *handler) reconnect(*websocket.Conn) {}
