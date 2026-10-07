package users

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	. "github.com/onsi/gomega"
)

func TestTrackKeepsOneSocketPerIP(t *testing.T) {
	g := NewWithT(t)

	people := New().(*tracker)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			OriginPatterns: []string{"127.0.0.1:*", "localhost:*"},
		})
		if err != nil {
			return
		}
		people.Track(r, conn)
		for {
			_, _, err := conn.Read(r.Context())
			if err != nil {
				people.Disconnect(conn)
				return
			}
		}
	}))
	t.Cleanup(srv.Close)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	t.Cleanup(cancel)

	dial := func() *websocket.Conn {
		t.Helper()
		conn, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(srv.URL, "http"), nil)
		g.Expect(err).NotTo(HaveOccurred())
		return conn
	}

	first := dial()
	t.Cleanup(func() { first.CloseNow() })
	var tracked *websocket.Conn
	g.Eventually(func() *websocket.Conn {
		people.mu.Lock()
		defer people.mu.Unlock()
		return people.users["127.0.0.1"]
	}, time.Second, 10*time.Millisecond).ShouldNot(BeNil())
	people.mu.Lock()
	tracked = people.users["127.0.0.1"]
	people.mu.Unlock()

	second := dial()
	t.Cleanup(func() { second.CloseNow() })
	g.Eventually(func() *websocket.Conn {
		people.mu.Lock()
		defer people.mu.Unlock()
		if len(people.users) != 1 {
			return nil
		}
		return people.users["127.0.0.1"]
	}, time.Second, 10*time.Millisecond).ShouldNot(Equal(tracked))

	second.CloseNow()
	g.Eventually(func() int {
		people.mu.Lock()
		defer people.mu.Unlock()
		return len(people.users)
	}, time.Second, 10*time.Millisecond).Should(Equal(0))
}
