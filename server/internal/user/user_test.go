package user

import (
	"math"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Colour", func() {
	It("gives one id one colour", func() {
		Expect(New("player-one").Colour()).To(Equal(New("player-one").Colour()))
		Expect(New("player-one").ID()).To(Equal("player-one"))
		Expect(New("player-one").Colour()).NotTo(Equal(New("player-two").Colour()))
	})

	It("keeps the palette apart and readable", func() {
		paper := [3]int{0xF9, 0xF7, 0xF7}
		mist := [3]int{0xDB, 0xE2, 0xEF}
		Expect(len(palette)).To(BeNumerically(">=", 50))
		Expect(palette[:3]).To(Equal([]string{"#112D4E", "#3F72AF", "#7AA0CE"}))
		Expect(closest(palette)).To(BeNumerically(">=", 36))
		for _, colour := range palette {
			got := parseHex(colour)
			Expect(contrast(got, paper)).To(BeNumerically(">=", 2))
			Expect(contrast(got, mist)).To(BeNumerically(">=", 2))
		}
	})
})

func closest(colours []string) float64 {
	GinkgoHelper()
	best := 1e9
	for i := range colours {
		for j := i + 1; j < len(colours); j++ {
			d := distance(parseHex(colours[i]), parseHex(colours[j]))
			if d < best {
				best = d
			}
		}
	}
	return best
}

func parseHex(colour string) [3]int {
	GinkgoHelper()
	Expect(colour).To(HaveLen(7))
	r, err := strconv.ParseInt(colour[1:3], 16, 0)
	Expect(err).NotTo(HaveOccurred())
	g, err := strconv.ParseInt(colour[3:5], 16, 0)
	Expect(err).NotTo(HaveOccurred())
	b, err := strconv.ParseInt(colour[5:7], 16, 0)
	Expect(err).NotTo(HaveOccurred())
	return [3]int{int(r), int(g), int(b)}
}

func distance(a, b [3]int) float64 {
	dr := float64(a[0] - b[0])
	dg := float64(a[1] - b[1])
	db := float64(a[2] - b[2])
	return math.Sqrt(dr*dr + dg*dg + db*db)
}

func contrast(a, b [3]int) float64 {
	lighter, darker := luminance(a), luminance(b)
	if darker > lighter {
		lighter, darker = darker, lighter
	}
	return (lighter + 0.05) / (darker + 0.05)
}

func luminance(rgb [3]int) float64 {
	channel := func(v int) float64 {
		c := float64(v) / 255
		if c <= 0.04045 {
			return c / 12.92
		}
		return math.Pow((c+0.055)/1.055, 2.4)
	}
	return 0.2126*channel(rgb[0]) + 0.7152*channel(rgb[1]) + 0.0722*channel(rgb[2])
}
