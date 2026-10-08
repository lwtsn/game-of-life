package users

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"game_of_life/server/internal/user/service"

	"github.com/coder/websocket"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("users", func() {
	It("keeps one socket per IP", func() {
		people := New(service.New()).(*tracker)
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
		DeferCleanup(srv.Close)

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		DeferCleanup(cancel)

		dial := func() *websocket.Conn {
			GinkgoHelper()
			conn, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(srv.URL, "http"), nil)
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(func() { conn.CloseNow() })
			return conn
		}

		dial()
		var tracked *websocket.Conn
		Eventually(func() *websocket.Conn {
			people.mu.Lock()
			defer people.mu.Unlock()
			return people.users["127.0.0.1"]
		}, time.Second, 10*time.Millisecond).ShouldNot(BeNil())
		people.mu.Lock()
		tracked = people.users["127.0.0.1"]
		people.mu.Unlock()

		second := dial()
		Eventually(func() *websocket.Conn {
			people.mu.Lock()
			defer people.mu.Unlock()
			if len(people.users) != 1 {
				return nil
			}
			return people.users["127.0.0.1"]
		}, time.Second, 10*time.Millisecond).ShouldNot(Equal(tracked))

		second.CloseNow()
		Eventually(func() int {
			people.mu.Lock()
			defer people.mu.Unlock()
			return len(people.users)
		}, time.Second, 10*time.Millisecond).Should(Equal(0))
	})
})
