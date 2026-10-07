package api

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

func (h *hub) track(ip string, conn *websocket.Conn) {
	h.mu.Lock()
	h.users[ip] = conn
	h.mu.Unlock()
}

func (h *hub) disconnect(conn *websocket.Conn) {
	h.mu.Lock()
	found := false
	for ip, current := range h.users {
		if current == conn {
			delete(h.users, ip)
			found = true
			break
		}
	}
	h.mu.Unlock()
	if !found {
		return
	}
	_ = conn.Close(websocket.StatusGoingAway, "")
	h.drop(conn)
	h.reconnect(conn)
}

// drop and reconnect are stubs. Who gets dropped, and how a client reconnects, comes later.
func (h *hub) drop(*websocket.Conn)      {}
func (h *hub) reconnect(*websocket.Conn) {}
