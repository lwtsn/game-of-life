package users

import (
	"net"
	"net/http"
	"slices"
	"sync"

	"github.com/coder/websocket"
)

// Tracker stores one socket per client IP.
type Tracker interface {
	Track(*http.Request, *websocket.Conn)
	Disconnect(*websocket.Conn)
	Conns() []*websocket.Conn
	Colours() []string
}

type tracker struct {
	mu    sync.Mutex
	users map[string]*websocket.Conn
}

func New() Tracker {
	return &tracker{users: make(map[string]*websocket.Conn)}
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func (t *tracker) Track(r *http.Request, conn *websocket.Conn) {
	t.mu.Lock()
	t.users[canonicalIP(clientIP(r))] = conn
	t.mu.Unlock()
}

func (t *tracker) Colours() []string {
	t.mu.Lock()
	defer t.mu.Unlock()
	colours := make([]string, 0, len(t.users))
	for ip := range t.users {
		colours = append(colours, Colour(ip))
	}
	slices.Sort(colours)
	return colours
}

func (t *tracker) Conns() []*websocket.Conn {
	t.mu.Lock()
	defer t.mu.Unlock()
	conns := make([]*websocket.Conn, 0, len(t.users))
	for _, conn := range t.users {
		conns = append(conns, conn)
	}
	return conns
}

func (t *tracker) Disconnect(conn *websocket.Conn) {
	t.mu.Lock()
	found := false
	for ip, current := range t.users {
		if current == conn {
			delete(t.users, ip)
			found = true
			break
		}
	}
	t.mu.Unlock()
	if !found {
		return
	}
	_ = conn.Close(websocket.StatusGoingAway, "")
	t.drop(conn)
	t.reconnect(conn)
}

// drop and reconnect are stubs. Who gets dropped, and how a client reconnects, comes later.
func (t *tracker) drop(*websocket.Conn)      {}
func (t *tracker) reconnect(*websocket.Conn) {}
