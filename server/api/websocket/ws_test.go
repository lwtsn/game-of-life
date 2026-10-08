package websocket

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"time"

	"game_of_life/server/api/users"
	"game_of_life/server/internal/grid/mocks"
	"game_of_life/server/internal/grid/source"
	"game_of_life/server/internal/user/service"

	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type fakeFrame struct {
	width  int
	height int
	cells  []int
}

func (f fakeFrame) Width() int   { return f.width }
func (f fakeFrame) Height() int  { return f.height }
func (f fakeFrame) Cells() []int { return f.cells }
func (f fakeFrame) ToJson() ([]byte, error) {
	return source.Encode(f)
}

func testServer(board *mocks.MockGrid, people users.Tracker) *httptest.Server {
	GinkgoHelper()
	h := New(board, people)
	engine := gin.New()
	engine.GET("/ws", h.Serve)
	srv := httptest.NewServer(engine)
	DeferCleanup(srv.Close)
	return srv
}

var _ = Describe("websocket", func() {
	It("sends the current board on connect", func() {
		frame := fakeFrame{width: 2, height: 2, cells: []int{1, 0, 0, 1}}
		board := mocks.NewMockGrid(GinkgoT())
		board.EXPECT().Current().Return(frame).Once()
		people := users.New(service.New())
		srv := testServer(board, people)

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		DeferCleanup(cancel)

		conn, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(srv.URL, "http")+"/ws", nil)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(conn.CloseNow)

		data := readBoard(conn, ctx)
		want, err := frame.ToJson()
		Expect(err).NotTo(HaveOccurred())
		Expect(data).To(Equal(want))
		Expect(people.Conns()).To(HaveLen(1))

		var presence struct {
			Colours []string `json:"colours"`
		}
		Expect(json.Unmarshal(readMessage(conn, ctx), &presence)).To(Succeed())
		Expect(presence.Colours).To(Equal([]string{service.New().Colour("127.0.0.1")}))
	})

	It("forwards later board updates", func() {
		first := fakeFrame{width: 2, height: 2, cells: []int{1, 0, 0, 0}}
		second := fakeFrame{width: 2, height: 2, cells: []int{0, 1, 0, 0}}
		secondJSON, err := second.ToJson()
		Expect(err).NotTo(HaveOccurred())

		updates := make(chan []byte, 1)
		var stream <-chan []byte = updates

		board := mocks.NewMockGrid(GinkgoT())
		board.EXPECT().Current().Return(first).Once()
		board.EXPECT().Updates().Return(stream)

		people := users.New(service.New())
		h := New(board, people)
		engine := gin.New()
		engine.GET("/ws", h.Serve)
		srv := httptest.NewServer(engine)
		DeferCleanup(srv.Close)

		ctx, cancel := context.WithCancel(context.Background())
		DeferCleanup(cancel)
		go h.Run(ctx)

		dialCtx, dialCancel := context.WithTimeout(context.Background(), 2*time.Second)
		DeferCleanup(dialCancel)

		conn, _, err := websocket.Dial(dialCtx, "ws"+strings.TrimPrefix(srv.URL, "http")+"/ws", nil)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(conn.CloseNow)

		data := readBoard(conn, dialCtx)
		want, err := first.ToJson()
		Expect(err).NotTo(HaveOccurred())
		Expect(data).To(Equal(want))

		updates <- secondJSON

		Expect(readBoard(conn, dialCtx)).To(Equal(secondJSON))
	})
})

func readMessage(conn *websocket.Conn, ctx context.Context) []byte {
	GinkgoHelper()
	_, data, err := conn.Read(ctx)
	Expect(err).NotTo(HaveOccurred())
	return data
}

func readBoard(conn *websocket.Conn, ctx context.Context) []byte {
	GinkgoHelper()
	for {
		data := readMessage(conn, ctx)
		var body map[string]any
		Expect(json.Unmarshal(data, &body)).To(Succeed())
		if _, ok := body["colours"]; ok {
			if _, grid := body["width"]; !grid {
				continue
			}
		}
		return data
	}
}
