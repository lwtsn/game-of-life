package user

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
)

var _ = Describe("service", func() {
	var svc Service

	BeforeEach(func() {
		app := fx.New(Module, fx.Populate(&svc), fx.NopLogger)
		Expect(app.Err()).NotTo(HaveOccurred())
	})

	It("rejects an id that is not a session", func() {
		_, ok := svc.Join("")
		Expect(ok).To(BeFalse())
		_, ok = svc.Join("127.0.0.1")
		Expect(ok).To(BeFalse())
		_, ok = svc.ByID("short")
		Expect(ok).To(BeFalse())
	})

	It("keeps one colour for a session across a refresh", func() {
		first, ok := svc.Join("player-one")
		Expect(ok).To(BeTrue())
		Expect(first.ID()).To(Equal("player-one"))
		Expect(first.Colour()).To(Equal("#112D4E"))
		Expect(svc.Colours()).To(Equal([]string{"#112D4E"}))

		again, ok := svc.Join("player-one")
		Expect(ok).To(BeTrue())
		Expect(again.Colour()).To(Equal(first.Colour()))
		Expect(svc.Colours()).To(Equal([]string{first.Colour()}))

		second, ok := svc.Join("player-two")
		Expect(ok).To(BeTrue())
		Expect(second.Colour()).To(Equal("#3F72AF"))
		Expect(svc.Colours()).To(Equal([]string{"#112D4E", "#3F72AF"}))

		svc.Leave("player-one")
		Expect(svc.Colours()).To(Equal([]string{second.Colour()}))

		back, ok := svc.Join("player-one")
		Expect(ok).To(BeTrue())
		Expect(back.Colour()).To(Equal("#112D4E"))
		Expect(svc.Colours()).To(Equal([]string{"#112D4E", "#3F72AF"}))
	})

	It("does not list a session until it joins", func() {
		person, ok := svc.ByID("player-one")
		Expect(ok).To(BeTrue())
		Expect(person.Colour()).To(Equal("#112D4E"))
		Expect(svc.Colours()).To(BeEmpty())

		joined, ok := svc.Join("player-one")
		Expect(ok).To(BeTrue())
		Expect(joined.Colour()).To(Equal(person.Colour()))
		Expect(svc.Colours()).To(Equal([]string{person.Colour()}))
	})

	It("rolls to the next colour nobody connected is using", func() {
		_, ok := svc.Join("player-one")
		Expect(ok).To(BeTrue())
		_, ok = svc.Join("player-two")
		Expect(ok).To(BeTrue())

		rolled, ok := svc.Recolour("player-one")
		Expect(ok).To(BeTrue())
		Expect(rolled.Colour()).To(Equal("#7AA0CE"))

		current, ok := svc.ByID("player-one")
		Expect(ok).To(BeTrue())
		Expect(current.Colour()).To(Equal("#7AA0CE"))
		Expect(svc.Colours()).To(Equal([]string{"#3F72AF", "#7AA0CE"}))

		_, ok = svc.Recolour("missing01")
		Expect(ok).To(BeFalse())
	})
})
