package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"game_of_life/server/api/controls"
	"game_of_life/server/api/users"
	"game_of_life/server/api/websocket"
	"game_of_life/server/internal/grid/mocks"
	"game_of_life/server/internal/grid/source"

	coderws "github.com/coder/websocket"
	"github.com/gin-gonic/gin"
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
		sockets:  websocket.New(board, users.New()),
		controls: controls.New(board),
	}
}

func TestRegisterConnectsTheRoutes(t *testing.T) {
	g := NewWithT(t)
	gin.SetMode(gin.TestMode)

	frame := fakeFrame{width: 2, height: 2, cells: []int{1, 0, 0, 1}}
	board := mocks.NewMockGrid(t)
	board.EXPECT().Current().Return(frame).Once()
	board.EXPECT().Start(mock.Anything).Once()

	h := testHandler(board)
	engine := gin.New()
	h.Register(engine)
	srv := httptest.NewServer(engine)
	t.Cleanup(srv.Close)

	res, err := http.Post(srv.URL+"/start", "application/json", nil)
	g.Expect(err).NotTo(HaveOccurred())
	res.Body.Close()
	g.Expect(res.StatusCode).To(Equal(http.StatusNoContent))

	res, err = http.Post(srv.URL+"/layout", "application/json", strings.NewReader(`{"name":"glider"}`))
	g.Expect(err).NotTo(HaveOccurred())
	defer res.Body.Close()
	g.Expect(res.StatusCode).To(Equal(http.StatusOK))
	body, err := io.ReadAll(res.Body)
	g.Expect(err).NotTo(HaveOccurred())
	var got map[string]any
	g.Expect(json.Unmarshal(body, &got)).To(Succeed())
	g.Expect(got["layout"]).To(Equal("glider"))
	g.Expect(got["applied"]).To(Equal(false))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	t.Cleanup(cancel)
	conn, _, err := coderws.Dial(ctx, "ws"+strings.TrimPrefix(srv.URL, "http")+"/ws", nil)
	g.Expect(err).NotTo(HaveOccurred())
	t.Cleanup(func() { conn.CloseNow() })

	_, data, err := conn.Read(ctx)
	g.Expect(err).NotTo(HaveOccurred())
	want, err := frame.ToJson()
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(data).To(Equal(want))
}
