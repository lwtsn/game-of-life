package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"game_of_life/server/api/controls"
	"game_of_life/server/api/websocket"
	"game_of_life/server/internal/grid/mocks"
	"game_of_life/server/internal/grid/source"
	"game_of_life/server/internal/user"

	coderws "github.com/coder/websocket"
	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
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

func testHandler(board *mocks.MockGrid) *handler {
	return &handler{
		sockets:  websocket.New(board, user.NewService()),
		controls: controls.New(board),
	}
}

var _ = Describe("router", func() {
	It("connects the downstream routes", func() {
		frame := fakeFrame{width: 2, height: 2, cells: []int{1, 0, 0, 1}}
		board := mocks.NewMockGrid(GinkgoT())
		board.EXPECT().Current().Return(frame).Once()
		board.EXPECT().Start(mock.Anything).Once()

		h := testHandler(board)
		engine := gin.New()
		h.Register(engine)
		srv := httptest.NewServer(engine)
		DeferCleanup(srv.Close)

		res, err := http.Post(srv.URL+"/start", "application/json", nil)
		Expect(err).NotTo(HaveOccurred())
		res.Body.Close()
		Expect(res.StatusCode).To(Equal(http.StatusNoContent))

		res, err = http.Post(srv.URL+"/layout", "application/json", strings.NewReader(`{"name":"glider"}`))
		Expect(err).NotTo(HaveOccurred())
		defer res.Body.Close()
		Expect(res.StatusCode).To(Equal(http.StatusOK))
		body, err := io.ReadAll(res.Body)
		Expect(err).NotTo(HaveOccurred())
		var got map[string]any
		Expect(json.Unmarshal(body, &got)).To(Succeed())
		Expect(got["layout"]).To(Equal("glider"))
		Expect(got["applied"]).To(Equal(false))

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		DeferCleanup(cancel)
		conn, _, err := coderws.Dial(ctx, "ws"+strings.TrimPrefix(srv.URL, "http")+"/ws", nil)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(conn.CloseNow)

		_, data, err := conn.Read(ctx)
		Expect(err).NotTo(HaveOccurred())
		want, err := frame.ToJson()
		Expect(err).NotTo(HaveOccurred())
		Expect(data).To(Equal(want))
	})
})
