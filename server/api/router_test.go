package api

import (
	"context"
	"net/http/httptest"
	"strings"
	"time"

	lifepb "game_of_life/server/gen/life/v1"
	"game_of_life/server/internal/grid"
	"game_of_life/server/internal/grid/mocks"
	"game_of_life/server/internal/grid/source"
	"game_of_life/server/internal/user"

	coderws "github.com/coder/websocket"
	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
)

type fakeFrame struct {
	width  int
	height int
	cells  []source.Cell
}

func (f fakeFrame) Width() int           { return f.width }
func (f fakeFrame) Height() int          { return f.height }
func (f fakeFrame) Cells() []source.Cell { return f.cells }
func (f fakeFrame) ToJson() ([]byte, error) {
	return source.Encode(f, 0)
}

func handlerFrom(board grid.Grid) Handler {
	GinkgoHelper()
	var h Handler
	app := fx.New(
		user.Module,
		Module,
		fx.Provide(func() grid.Grid { return board }),
		fx.Populate(&h),
		fx.NopLogger,
	)
	Expect(app.Err()).NotTo(HaveOccurred())
	return h
}

var _ = Describe("router", func() {
	It("connects the downstream routes", func() {
		frame := fakeFrame{width: 2, height: 2, cells: []source.Cell{{Alive: true}, {}, {}, {Alive: true}}}
		board := mocks.NewMockGrid(GinkgoT())
		board.EXPECT().Current().Return(frame).Once()
		board.EXPECT().Clock().Return(true, int(lifepb.PaceBound_PACE_BOUND_MIN)).Once()

		h := handlerFrom(board)
		engine := gin.New()
		h.Register(engine)
		srv := httptest.NewServer(engine)
		DeferCleanup(srv.Close)

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
