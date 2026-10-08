package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"game_of_life/server/internal/user"

	"github.com/coder/websocket"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
)

type socketCell struct {
	Alive  bool   `json:"alive"`
	IP     string `json:"ip"`
	Colour string `json:"colour"`
}

type socketFrame struct {
	Width  int          `json:"width"`
	Height int          `json:"height"`
	Cells  []socketCell `json:"cells"`
}

var _ = Describe("server", func() {
	It("starts and stops", func() {
		app := fx.New(
			module(),
			fx.Replace(config{addr: "127.0.0.1:0"}),
			fx.NopLogger,
		)
		Expect(app.Err()).NotTo(HaveOccurred())

		ctx := context.Background()
		Expect(app.Start(ctx)).To(Succeed())
		Expect(app.Stop(ctx)).To(Succeed())
	})

	It("serves the board", func() {
		var srv *http.Server
		app := fx.New(
			module(),
			fx.Replace(config{addr: "127.0.0.1:0"}),
			fx.Populate(&srv),
			fx.NopLogger,
		)
		Expect(app.Err()).NotTo(HaveOccurred())
		Expect(app.Start(context.Background())).To(Succeed())
		DeferCleanup(func() {
			stopCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			Expect(app.Stop(stopCtx)).To(Succeed())
		})

		Expect(srv).NotTo(BeNil())
		Expect(srv.Addr).NotTo(HaveSuffix(":0"))

		dialCtx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
		DeferCleanup(cancel)

		conn, _, err := websocket.Dial(dialCtx, "ws://"+srv.Addr+"/ws", nil)
		Expect(err).NotTo(HaveOccurred())
		conn.SetReadLimit(1 << 20)
		DeferCleanup(conn.CloseNow)

		first := readSocketFrame(conn, dialCtx)
		second := readSocketFrame(conn, dialCtx)
		expectFullGrid(first)
		expectFullGrid(second)
		Expect(first.Cells).To(Equal(make([]socketCell, 80*50)))
		Expect(second.Cells).To(Equal(first.Cells))
	})

	It("sends a placed cell to a second connection", func() {
		var srv *http.Server
		app := fx.New(
			module(),
			fx.Replace(config{addr: "127.0.0.1:0"}),
			fx.Populate(&srv),
			fx.NopLogger,
		)
		Expect(app.Err()).NotTo(HaveOccurred())
		Expect(app.Start(context.Background())).To(Succeed())
		DeferCleanup(func() {
			stopCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			Expect(app.Stop(stopCtx)).To(Succeed())
		})

		dialCtx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
		DeferCleanup(cancel)

		conn, _, err := websocket.Dial(dialCtx, "ws://"+srv.Addr+"/ws", nil)
		Expect(err).NotTo(HaveOccurred())
		conn.SetReadLimit(1 << 20)
		DeferCleanup(conn.CloseNow)
		_ = readSocketFrame(conn, dialCtx)

		person := user.New("127.0.0.1")
		places := [][2]int{{0, 0}, {1, 0}, {0, 1}, {1, 1}}
		for _, xy := range places {
			payload, err := json.Marshal(struct {
				X int `json:"x"`
				Y int `json:"y"`
			}{X: xy[0], Y: xy[1]})
			Expect(err).NotTo(HaveOccurred())
			Expect(conn.Write(dialCtx, websocket.MessageText, payload)).To(Succeed())
		}

		placed := readUntil(conn, dialCtx, func(frame socketFrame) bool {
			if len(frame.Cells) != 80*50 {
				return false
			}
			for _, xy := range places {
				cell := frame.Cells[xy[1]*80+xy[0]]
				if !cell.Alive || cell.Colour != person.Colour() || cell.IP != person.IP() {
					return false
				}
			}
			return true
		})
		Expect(placed.Cells[0].Colour).To(Equal(person.Colour()))

		other, _, err := websocket.Dial(dialCtx, "ws://"+srv.Addr+"/ws", nil)
		Expect(err).NotTo(HaveOccurred())
		other.SetReadLimit(1 << 20)
		DeferCleanup(other.CloseNow)

		seen := readSocketFrame(other, dialCtx)
		for _, xy := range places {
			cell := seen.Cells[xy[1]*80+xy[0]]
			Expect(cell.Alive).To(BeTrue())
			Expect(cell.Colour).To(Equal(person.Colour()))
			Expect(cell.IP).To(Equal("127.0.0.1"))
		}
	})

})

func readSocketFrame(conn *websocket.Conn, ctx context.Context) socketFrame {
	GinkgoHelper()
	for {
		_, data, err := conn.Read(ctx)
		Expect(err).NotTo(HaveOccurred())

		var got socketFrame
		Expect(json.Unmarshal(data, &got)).To(Succeed())
		if got.Width == 0 {
			continue
		}
		return got
	}
}

func readUntil(conn *websocket.Conn, ctx context.Context, match func(socketFrame) bool) socketFrame {
	GinkgoHelper()
	for {
		frame := readSocketFrame(conn, ctx)
		if match(frame) {
			return frame
		}
	}
}

func expectFullGrid(frame socketFrame) {
	GinkgoHelper()
	Expect(frame.Width).To(Equal(80))
	Expect(frame.Height).To(Equal(50))
	Expect(frame.Cells).To(HaveLen(80 * 50))
}
