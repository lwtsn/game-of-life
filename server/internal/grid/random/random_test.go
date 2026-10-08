package random

import (
	"game_of_life/server/internal/grid/source"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
)

var _ = Describe("random source", func() {
	var src source.Source

	BeforeEach(func() {
		app := fx.New(Module, fx.Populate(&src), fx.NopLogger)
		Expect(app.Err()).NotTo(HaveOccurred())
	})

	It("fills the whole grid", func() {
		first := src.Next(nil)
		second := src.Next(nil)

		Expect(first.Width()).To(Equal(cols))
		Expect(first.Height()).To(Equal(rows))
		Expect(first.Cells()).To(HaveLen(cols * rows))
		alive := 0
		dead := 0
		for _, cell := range first.Cells() {
			Expect(cell.User).To(BeNil())
			if cell.Alive {
				alive++
			} else {
				dead++
			}
		}
		Expect(alive).To(BeNumerically(">", 0))
		Expect(dead).To(BeNumerically(">", 0))
		Expect(first.Cells()).NotTo(Equal(second.Cells()))
	})
})
