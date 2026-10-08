package layout

import (
	lifepb "game_of_life/server/gen/life/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
)

func catalogueFrom() Catalogue {
	GinkgoHelper()
	var shapes Catalogue
	app := fx.New(Module, fx.Populate(&shapes), fx.NopLogger)
	Expect(app.Err()).NotTo(HaveOccurred())
	return shapes
}

var _ = Describe("Catalogue", func() {
	It("loads the four shapes from the text catalogue", func() {
		shapes := catalogueFrom()

		_, ok := shapes.ByPattern(lifepb.Pattern_PATTERN_UNSPECIFIED)
		Expect(ok).To(BeFalse())

		block, ok := shapes.ByPattern(lifepb.Pattern_PATTERN_BLOCK)
		Expect(ok).To(BeTrue())
		Expect(block.Label).To(Equal("Block"))
		Expect(block.Width).To(Equal(2))
		Expect(block.Height).To(Equal(2))
		Expect(block.Cells).To(Equal([]Cell{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 0, Y: 1}, {X: 1, Y: 1}}))

		blinker, ok := shapes.ByPattern(lifepb.Pattern_PATTERN_BLINKER)
		Expect(ok).To(BeTrue())
		Expect(blinker.Label).To(Equal("Blinker"))
		Expect(blinker.Width).To(Equal(3))
		Expect(blinker.Height).To(Equal(2))
		Expect(blinker.Cells).To(Equal([]Cell{{X: 0, Y: 1}, {X: 1, Y: 1}, {X: 2, Y: 1}}))

		glider, ok := shapes.ByPattern(lifepb.Pattern_PATTERN_GLIDER)
		Expect(ok).To(BeTrue())
		Expect(glider.Label).To(Equal("Glider"))
		Expect(glider.Width).To(Equal(3))
		Expect(glider.Height).To(Equal(3))
		Expect(glider.Cells).To(Equal([]Cell{{X: 1, Y: 0}, {X: 2, Y: 1}, {X: 0, Y: 2}, {X: 1, Y: 2}, {X: 2, Y: 2}}))

		beacon, ok := shapes.ByPattern(lifepb.Pattern_PATTERN_BEACON)
		Expect(ok).To(BeTrue())
		Expect(beacon.Label).To(Equal("Beacon"))
		Expect(beacon.Width).To(Equal(4))
		Expect(beacon.Height).To(Equal(4))
		Expect(beacon.Cells).To(Equal([]Cell{
			{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 0, Y: 1}, {X: 1, Y: 1},
			{X: 2, Y: 2}, {X: 3, Y: 2}, {X: 2, Y: 3}, {X: 3, Y: 3},
		}))
	})

	It("keeps a random origin inside the board", func() {
		for range 40 {
			x, y := randomOrigin(80, 50, 4, 4)
			Expect(x).To(BeNumerically(">=", 0))
			Expect(y).To(BeNumerically(">=", 0))
			Expect(x).To(BeNumerically("<=", 76))
			Expect(y).To(BeNumerically("<=", 46))
		}
	})
})
