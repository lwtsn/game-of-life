package websocket

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"game_of_life/server/internal/grid"
	"game_of_life/server/internal/grid/mocks"
	"game_of_life/server/internal/user"

	"github.com/coder/websocket"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
)

var _ = Describe("clients", func() {
	It("keeps every socket for the same session", func() {
		board := mocks.NewMockGrid(GinkgoT())
		var api Handler
		app := fx.New(
			user.Module,
			Module,
			fx.Provide(func() grid.Grid { return board }),
			fx.Populate(&api),
			fx.NopLogger,
		)
		Expect(app.Err()).NotTo(HaveOccurred())
		people := api.(*handler)
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
				OriginPatterns: []string{"127.0.0.1:*", "localhost:*"},
			})
			if err != nil {
				return
			}
			people.track(r, conn)
			for {
				_, _, err := conn.Read(r.Context())
				if err != nil {
					people.disconnect(conn)
					return
				}
			}
		}))
		DeferCleanup(srv.Close)

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		DeferCleanup(cancel)

		dial := func(session string) *websocket.Conn {
			GinkgoHelper()
			conn, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(srv.URL, "http")+"/?session="+session, nil)
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(func() { conn.CloseNow() })
			return conn
		}

		first := dial("player-one")
		Eventually(func() int {
			people.mu.Lock()
			defer people.mu.Unlock()
			return len(people.clients)
		}, time.Second, 10*time.Millisecond).Should(Equal(1))

		second := dial("player-one")
		Eventually(func() int {
			people.mu.Lock()
			defer people.mu.Unlock()
			return len(people.clients)
		}, time.Second, 10*time.Millisecond).Should(Equal(2))
		Expect(people.svc.Colours()).To(HaveLen(1))

		second.CloseNow()
		Eventually(func() int {
			people.mu.Lock()
			defer people.mu.Unlock()
			return len(people.clients)
		}, time.Second, 10*time.Millisecond).Should(Equal(1))
		Expect(people.svc.Colours()).To(HaveLen(1))

		first.CloseNow()
		Eventually(func() int {
			people.mu.Lock()
			defer people.mu.Unlock()
			return len(people.clients)
		}, time.Second, 10*time.Millisecond).Should(Equal(0))
		Expect(people.svc.Colours()).To(BeEmpty())
	})
})
