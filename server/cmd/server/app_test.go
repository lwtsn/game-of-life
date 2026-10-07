package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/coder/websocket"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
)

type socketFrame struct {
	Width  int   `json:"width"`
	Height int   `json:"height"`
	Cells  []int `json:"cells"`
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

	It("serves a live grid", func() {
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
		DeferCleanup(conn.CloseNow)

		first := readSocketFrame(conn, dialCtx)
		second := readSocketFrame(conn, dialCtx)
		expectFullGrid(first)
		expectFullGrid(second)
		Expect(first.Cells).NotTo(Equal(second.Cells))
	})
})

func readSocketFrame(conn *websocket.Conn, ctx context.Context) socketFrame {
	GinkgoHelper()
	_, data, err := conn.Read(ctx)
	Expect(err).NotTo(HaveOccurred())

	var got socketFrame
	Expect(json.Unmarshal(data, &got)).To(Succeed())
	return got
}

func expectFullGrid(frame socketFrame) {
	GinkgoHelper()
	Expect(frame.Width).To(Equal(80))
	Expect(frame.Height).To(Equal(50))
	Expect(frame.Cells).To(HaveLen(80 * 50))
	Expect(frame.Cells).To(ContainElement(0))
	Expect(frame.Cells).To(ContainElement(1))
	for _, cell := range frame.Cells {
		Expect(cell).To(BeElementOf(0, 1))
	}
}
