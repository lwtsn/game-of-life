package controls

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"game_of_life/server/internal/grid/mocks"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
)

func TestStartCallsTheGrid(t *testing.T) {
	g := NewWithT(t)
	gin.SetMode(gin.TestMode)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	board := mocks.NewMockGrid(t)
	board.EXPECT().Start(mock.MatchedBy(func(got context.Context) bool {
		return got == ctx
	})).Once()

	h := New(board)
	h.Bind(ctx)

	engine := gin.New()
	engine.POST("/start", h.Start)
	srv := httptest.NewServer(engine)
	t.Cleanup(srv.Close)

	res, err := http.Post(srv.URL+"/start", "application/json", nil)
	g.Expect(err).NotTo(HaveOccurred())
	defer res.Body.Close()
	g.Expect(res.StatusCode).To(Equal(http.StatusNoContent))
}

func TestLayoutIsAStub(t *testing.T) {
	g := NewWithT(t)
	gin.SetMode(gin.TestMode)

	board := mocks.NewMockGrid(t)
	h := New(board)

	engine := gin.New()
	engine.POST("/layout", h.Layout)
	srv := httptest.NewServer(engine)
	t.Cleanup(srv.Close)

	res, err := http.Post(srv.URL+"/layout", "application/json", strings.NewReader(`{"name":"glider"}`))
	g.Expect(err).NotTo(HaveOccurred())
	defer res.Body.Close()
	g.Expect(res.StatusCode).To(Equal(http.StatusOK))

	body, err := io.ReadAll(res.Body)
	g.Expect(err).NotTo(HaveOccurred())

	var got map[string]any
	g.Expect(json.Unmarshal(body, &got)).To(Succeed())
	g.Expect(got["layout"]).To(Equal("glider"))
	g.Expect(got["applied"]).To(Equal(false))
}
