package controls

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"game_of_life/server/internal/grid"
	"game_of_life/server/internal/grid/mocks"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	"go.uber.org/fx"
)

func handlerFrom(board grid.Grid) Handler {
	GinkgoHelper()
	var h Handler
	app := fx.New(
		Module,
		fx.Provide(func() grid.Grid { return board }),
		fx.Populate(&h),
		fx.NopLogger,
	)
	Expect(app.Err()).NotTo(HaveOccurred())
	return h
}

var _ = Describe("controls", func() {
	It("starts the grid with the bound context", func() {
		ctx, cancel := context.WithCancel(context.Background())
		DeferCleanup(cancel)

		board := mocks.NewMockGrid(GinkgoT())
		board.EXPECT().Start(mock.MatchedBy(func(got context.Context) bool {
			return got == ctx
		})).Once()

		h := handlerFrom(board)
		h.Bind(ctx)

		engine := gin.New()
		engine.POST("/start", h.Start)
		srv := httptest.NewServer(engine)
		DeferCleanup(srv.Close)

		res, err := http.Post(srv.URL+"/start", "application/json", nil)
		Expect(err).NotTo(HaveOccurred())
		defer res.Body.Close()
		Expect(res.StatusCode).To(Equal(http.StatusNoContent))
	})

	It("accepts a layout name and does not apply it", func() {
		board := mocks.NewMockGrid(GinkgoT())
		h := handlerFrom(board)

		engine := gin.New()
		engine.POST("/layout", h.Layout)
		srv := httptest.NewServer(engine)
		DeferCleanup(srv.Close)

		res, err := http.Post(srv.URL+"/layout", "application/json", strings.NewReader(`{"name":"glider"}`))
		Expect(err).NotTo(HaveOccurred())
		defer res.Body.Close()
		Expect(res.StatusCode).To(Equal(http.StatusOK))

		body, err := io.ReadAll(res.Body)
		Expect(err).NotTo(HaveOccurred())

		var got map[string]any
		Expect(json.Unmarshal(body, &got)).To(Succeed())
		Expect(got["layout"]).To(Equal("glider"))
		Expect(got["applied"]).To(Equal(false))
	})
})
