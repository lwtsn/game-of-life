package websocket

import (
	"net"
	"net/http"

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
	h.clients[person.IP()] = conn
	h.mu.Unlock()
}

func (h *handler) conns() []*websocket.Conn {
	h.mu.Lock()
	defer h.mu.Unlock()
	conns := make([]*websocket.Conn, 0, len(h.clients))
	for _, conn := range h.clients {
		conns = append(conns, conn)
	}
	return conns
}

func (h *handler) disconnect(conn *websocket.Conn) {
	h.mu.Lock()
	ip := ""
	found := false
	for key, current := range h.clients {
		if current == conn {
			delete(h.clients, key)
			ip = key
			found = true
			break
		}
	}
	h.mu.Unlock()
	if !found {
		return
	}
	h.svc.Leave(ip)
	_ = conn.Close(websocket.StatusGoingAway, "")
	h.drop(conn)
	h.reconnect(conn)
}

// drop and reconnect are stubs. Who gets dropped, and how a client reconnects, comes later.
func (h *handler) drop(*websocket.Conn)      {}
func (h *handler) reconnect(*websocket.Conn) {}
