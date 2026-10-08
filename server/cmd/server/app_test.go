package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	lifepb "game_of_life/server/gen/life/v1"
	"game_of_life/server/internal/layout"

	"github.com/coder/websocket"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	boardWidth  = int(lifepb.GridSize_GRID_SIZE_WIDTH)
	boardHeight = int(lifepb.GridSize_GRID_SIZE_HEIGHT)
)

type socketCell struct {
	Alive  bool   `json:"alive"`
	ID     string `json:"id"`
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
		Expect(first.Cells).To(Equal(make([]socketCell, boardWidth*boardHeight)))
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

		conn, _, err := websocket.Dial(dialCtx, "ws://"+srv.Addr+"/ws?session=player-one", nil)
		Expect(err).NotTo(HaveOccurred())
		conn.SetReadLimit(1 << 20)
		DeferCleanup(conn.CloseNow)
		_ = readSocketFrame(conn, dialCtx)

		places := [][2]int{{0, 0}, {1, 0}, {0, 1}, {1, 1}}
		for _, xy := range places {
			writePoint(conn, dialCtx, xy[0], xy[1])
		}

		placed := readUntil(conn, dialCtx, func(frame socketFrame) bool {
			if len(frame.Cells) != boardWidth*boardHeight {
				return false
			}
			for _, xy := range places {
				cell := frame.Cells[xy[1]*boardWidth+xy[0]]
				if !cell.Alive || cell.Colour != "#112D4E" || cell.ID != "player-one" {
					return false
				}
			}
			return true
		})
		Expect(placed.Cells[0].Colour).To(Equal("#112D4E"))

		other, _, err := websocket.Dial(dialCtx, "ws://"+srv.Addr+"/ws?session=player-one", nil)
		Expect(err).NotTo(HaveOccurred())
		other.SetReadLimit(1 << 20)
		DeferCleanup(other.CloseNow)

		seen := readSocketFrame(other, dialCtx)
		for _, xy := range places {
			cell := seen.Cells[xy[1]*boardWidth+xy[0]]
			Expect(cell.Alive).To(BeTrue())
			Expect(cell.Colour).To(Equal("#112D4E"))
			Expect(cell.ID).To(Equal("player-one"))
		}
	})

	It("sends the whole board when a birth takes its parents' colour", func() {
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

		dialCtx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
		DeferCleanup(cancel)

		conn, _, err := websocket.Dial(dialCtx, "ws://"+srv.Addr+"/ws?session=player-one", nil)
		Expect(err).NotTo(HaveOccurred())
		conn.SetReadLimit(1 << 20)
		DeferCleanup(conn.CloseNow)
		_ = readSocketFrame(conn, dialCtx)

		for _, xy := range [][2]int{{10, 10}, {11, 10}, {12, 10}} {
			writePoint(conn, dialCtx, xy[0], xy[1])
		}

		born := readUntil(conn, dialCtx, func(frame socketFrame) bool {
			if len(frame.Cells) != boardWidth*boardHeight {
				return false
			}
			birth := frame.Cells[9*boardWidth+11]
			return birth.Alive && birth.Colour == "#112D4E" && birth.ID == ""
		})
		expectFullGrid(born)
		survivor := born.Cells[10*boardWidth+11]
		Expect(survivor.Alive).To(BeTrue())
		Expect(survivor.Colour).To(Equal("#112D4E"))
		Expect(survivor.ID).To(Equal("player-one"))
	})

	It("allows Connect from the local page and refuses another origin", func() {
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

		ask := func(origin string) *http.Response {
			GinkgoHelper()
			req, err := http.NewRequest(http.MethodOptions, "http://"+srv.Addr+"/life.v1.LayoutService/Place", nil)
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Origin", origin)
			req.Header.Set("Access-Control-Request-Method", "POST")
			res, err := http.DefaultClient.Do(req)
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(res.Body.Close)
			return res
		}

		local := ask("http://localhost:5173")
		Expect(local.StatusCode).To(Equal(http.StatusNoContent))
		Expect(local.Header.Get("Access-Control-Allow-Origin")).To(Equal("http://localhost:5173"))
		Expect(local.Header.Get("Access-Control-Allow-Headers")).To(ContainSubstring("Connect-Protocol-Version"))

		other := ask("https://example.com")
		Expect(other.Header.Get("Access-Control-Allow-Origin")).To(BeEmpty())
	})

	It("stamps a block through Connect and sends it on the socket", func() {
		var srv *http.Server
		app := fx.New(
			module(),
			fx.Replace(config{addr: "127.0.0.1:0"}),
			fx.Replace(layout.Origin(func(int, int, int, int) (int, int) { return 4, 5 })),
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

		conn, _, err := websocket.Dial(dialCtx, "ws://"+srv.Addr+"/ws?session=player-one", nil)
		Expect(err).NotTo(HaveOccurred())
		conn.SetReadLimit(1 << 20)
		DeferCleanup(conn.CloseNow)
		_ = readSocketFrame(conn, dialCtx)

		rejected, rejectedBody := postPlace(srv, `{}`, "")
		Expect(rejected.StatusCode).To(Equal(http.StatusBadRequest))
		Expect(string(rejectedBody)).To(ContainSubstring("invalid_argument"))

		res, body := postPlace(srv, `{"pattern":"PATTERN_BLOCK"}`, "player-one")
		Expect(res.StatusCode).To(Equal(http.StatusOK))
		Expect(res.Header.Get("Content-Type")).To(ContainSubstring("application/json"))
		var placed lifepb.PlaceResponse
		Expect(protojson.Unmarshal(body, &placed)).To(Succeed())
		Expect(placed.GetApplied()).To(BeTrue())
		Expect(placed.GetPattern()).To(Equal(lifepb.Pattern_PATTERN_BLOCK))
		Expect(placed.GetOrigin().GetX()).To(Equal(int32(4)))
		Expect(placed.GetOrigin().GetY()).To(Equal(int32(5)))

		cells := [][2]int{{4, 5}, {5, 5}, {4, 6}, {5, 6}}
		seen := readUntil(conn, dialCtx, func(frame socketFrame) bool {
			if len(frame.Cells) != boardWidth*boardHeight {
				return false
			}
			for _, xy := range cells {
				cell := frame.Cells[xy[1]*boardWidth+xy[0]]
				if !cell.Alive || cell.Colour != "#112D4E" || cell.ID != "player-one" {
					return false
				}
			}
			return true
		})
		Expect(seen.Cells[5*boardWidth+4].Colour).To(Equal("#112D4E"))
	})

	It("clears the board and names the colour that asked", func() {
		var srv *http.Server
		app := fx.New(
			module(),
			fx.Replace(config{addr: "127.0.0.1:0"}),
			fx.Replace(layout.Origin(func(int, int, int, int) (int, int) { return 4, 5 })),
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

		conn, _, err := websocket.Dial(dialCtx, "ws://"+srv.Addr+"/ws?session=player-one", nil)
		Expect(err).NotTo(HaveOccurred())
		conn.SetReadLimit(1 << 20)
		DeferCleanup(conn.CloseNow)

		res, body := postPlace(srv, `{"pattern":"PATTERN_BLOCK"}`, "player-one")
		Expect(res.StatusCode).To(Equal(http.StatusOK))
		Expect(string(body)).To(ContainSubstring(`"applied":true`))

		_ = readUntil(conn, dialCtx, func(frame socketFrame) bool {
			if len(frame.Cells) != boardWidth*boardHeight {
				return false
			}
			cell := frame.Cells[5*boardWidth+4]
			return cell.Alive && cell.Colour == "#112D4E"
		})

		payload, err := protojson.Marshal(&lifepb.ClientMessage{
			Action: &lifepb.ClientMessage_ResetBoard{ResetBoard: true},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(conn.Write(dialCtx, websocket.MessageText, payload)).To(Succeed())

		var cleared bool
		var notice string
		for !cleared || notice == "" {
			_, data, err := conn.Read(dialCtx)
			Expect(err).NotTo(HaveOccurred())

			var raw lifepb.ServerMessage
			Expect(protojson.Unmarshal(data, &raw)).To(Succeed())
			if raw.GetType() == lifepb.MessageType_MESSAGE_TYPE_RESET && raw.GetColour() == "#112D4E" {
				notice = raw.GetColour()
			}
			if int(raw.GetWidth()) == boardWidth && len(raw.GetCells()) == boardWidth*boardHeight {
				alive := false
				for _, cell := range raw.GetCells() {
					if cell.GetAlive() {
						alive = true
						break
					}
				}
				if !alive {
					cleared = true
				}
			}
		}
		Expect(notice).To(Equal("#112D4E"))
	})
})

func writePoint(conn *websocket.Conn, ctx context.Context, x, y int) {
	GinkgoHelper()
	payload, err := protojson.Marshal(&lifepb.ClientMessage{
		Action: &lifepb.ClientMessage_Point{Point: &lifepb.Point{X: int32(x), Y: int32(y)}},
	})
	Expect(err).NotTo(HaveOccurred())
	Expect(conn.Write(ctx, websocket.MessageText, payload)).To(Succeed())
}

func postPlace(srv *http.Server, body, session string) (*http.Response, []byte) {
	GinkgoHelper()
	req, err := http.NewRequest(http.MethodPost, "http://"+srv.Addr+"/life.v1.LayoutService/Place", strings.NewReader(body))
	Expect(err).NotTo(HaveOccurred())
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Connect-Protocol-Version", "1")
	if session != "" {
		req.Header.Set("X-Session", session)
	}
	res, err := http.DefaultClient.Do(req)
	Expect(err).NotTo(HaveOccurred())
	defer res.Body.Close()
	payload, err := io.ReadAll(res.Body)
	Expect(err).NotTo(HaveOccurred())
	return res, payload
}

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
	Expect(frame.Width).To(Equal(boardWidth))
	Expect(frame.Height).To(Equal(boardHeight))
	Expect(frame.Cells).To(HaveLen(boardWidth * boardHeight))
}
