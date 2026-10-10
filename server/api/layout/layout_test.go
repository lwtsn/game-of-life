package layout

import (
	"context"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connectinprocess"
	"game_of_life/server/api/websocket"
	lifepb "game_of_life/server/gen/life/v1"
	lifepbconnect "game_of_life/server/gen/life/v1/lifepbconnect"
	"game_of_life/server/internal/grid"
	"game_of_life/server/internal/grid/life"
	"game_of_life/server/internal/user"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
)

func placeClient() (lifepbconnect.LayoutServiceClient, grid.Grid) {
	GinkgoHelper()
	var rpc *connect.Server
	var board grid.Grid
	app := fx.New(
		life.Module,
		grid.Module,
		user.Module,
		websocket.Module,
		Module,
		fx.Populate(&rpc, &board),
		fx.NopLogger,
	)
	Expect(app.Err()).NotTo(HaveOccurred())
	Expect(rpc).NotTo(BeNil())
	client := lifepbconnect.NewLayoutServiceClient(connect.NewClient(connectinprocess.New(rpc)))
	return client, board
}

func aliveCount(board grid.Grid) int {
	GinkgoHelper()
	alive := 0
	for _, cell := range board.Current().Cells() {
		if cell.Alive {
			alive++
		}
	}
	return alive
}

var _ = Describe("Place", func() {
	It("rejects an unknown pattern", func() {
		client, board := placeClient()
		_, err := client.Place(context.Background(), &lifepb.PlaceRequest{})
		Expect(err).To(HaveOccurred())
		Expect(connect.CodeOf(err)).To(Equal(connect.CodeInvalidArgument))
		Expect(aliveCount(board)).To(Equal(0))
	})

	It("does not stamp when the call has no session", func() {
		client, board := placeClient()
		_, err := client.Place(context.Background(), &lifepb.PlaceRequest{
			Pattern: lifepb.Pattern_PATTERN_BLOCK,
		})
		Expect(err).To(HaveOccurred())
		Expect(connect.CodeOf(err)).To(Equal(connect.CodeInvalidArgument))
		Expect(aliveCount(board)).To(Equal(0))
	})
})

var _ = Describe("ListPatterns", func() {
	It("returns the catalogue without a session", func() {
		client, _ := placeClient()
		res, err := client.ListPatterns(context.Background(), &lifepb.ListPatternsRequest{})
		Expect(err).NotTo(HaveOccurred())

		shapes := res.GetShapes()
		labels := make([]string, len(shapes))
		for i, shape := range shapes {
			labels[i] = shape.GetLabel()
		}
		Expect(labels).To(Equal([]string{"Block", "Blinker", "Glider", "Beacon"}))

		glider := shapes[2]
		Expect(glider.GetPattern()).To(Equal(lifepb.Pattern_PATTERN_GLIDER))
		cells := make([][2]int32, len(glider.GetCells()))
		for i, cell := range glider.GetCells() {
			cells[i] = [2]int32{cell.GetX(), cell.GetY()}
		}
		Expect(cells).To(Equal([][2]int32{{1, 0}, {2, 1}, {0, 2}, {1, 2}, {2, 2}}))
	})
})
