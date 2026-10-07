package main

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/coder/websocket"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
)

func TestAppStarts(t *testing.T) {
	g := NewWithT(t)

	app := fx.New(
		module(),
		fx.Replace(config{addr: "127.0.0.1:0"}),
		fx.NopLogger,
	)
	g.Expect(app.Err()).NotTo(HaveOccurred())

	ctx := context.Background()
	g.Expect(app.Start(ctx)).To(Succeed())
	g.Expect(app.Stop(ctx)).To(Succeed())
}

type socketFrame struct {
	Width  int   `json:"width"`
	Height int   `json:"height"`
	Cells  []int `json:"cells"`
}

func TestModuleWiresTheAPI(t *testing.T) {
	g := NewWithT(t)

	var srv *http.Server
	app := fx.New(
		module(),
		fx.Replace(config{addr: "127.0.0.1:0"}),
		fx.Populate(&srv),
		fx.NopLogger,
	)
	g.Expect(app.Err()).NotTo(HaveOccurred())
	g.Expect(app.Start(context.Background())).To(Succeed())
	t.Cleanup(func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		g.Expect(app.Stop(stopCtx)).To(Succeed())
	})

	g.Expect(srv).NotTo(BeNil())
	g.Expect(srv.Addr).NotTo(HaveSuffix(":0"))

	dialCtx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	t.Cleanup(cancel)

	conn, _, err := websocket.Dial(dialCtx, "ws://"+srv.Addr+"/ws", nil)
	g.Expect(err).NotTo(HaveOccurred())
	t.Cleanup(func() { conn.CloseNow() })

	first := readSocketFrame(g, conn, dialCtx)
	second := readSocketFrame(g, conn, dialCtx)
	expectFullGrid(g, first)
	expectFullGrid(g, second)
	g.Expect(first.Cells).NotTo(Equal(second.Cells))
}

func readSocketFrame(g *WithT, conn *websocket.Conn, ctx context.Context) socketFrame {
	_, data, err := conn.Read(ctx)
	g.Expect(err).NotTo(HaveOccurred())

	var got socketFrame
	g.Expect(json.Unmarshal(data, &got)).To(Succeed())
	return got
}

func expectFullGrid(g *WithT, frame socketFrame) {
	g.Expect(frame.Width).To(Equal(80))
	g.Expect(frame.Height).To(Equal(50))
	g.Expect(frame.Cells).To(HaveLen(80 * 50))
	g.Expect(frame.Cells).To(ContainElement(0))
	g.Expect(frame.Cells).To(ContainElement(1))
	for _, cell := range frame.Cells {
		g.Expect(cell).To(BeElementOf(0, 1))
	}
}
