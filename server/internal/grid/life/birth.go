package life

import (
	"fmt"
	"math"
)

// BirthColour mixes the three colours that made a new cell.
// The mix is done in OKLab, so it keeps the brightness of the parents.
func BirthColour(first, second, third string) (string, error) {
	var mixed lab
	for _, hex := range []string{first, second, third} {
		colour, err := parseSRGB(hex)
		if err != nil {
			return "", err
		}
		sample := toOKLab(colour)
		mixed.l += sample.l
		mixed.a += sample.a
		mixed.b += sample.b
	}
	mixed.l /= 3
	mixed.a /= 3
	mixed.b /= 3
	return formatSRGB(fromOKLab(mixed)), nil
}

type rgb struct{ r, g, b float64 }

// lab is one colour as OKLab sees it.
// l is how light it is. a runs from green to red. b runs from blue to yellow.
type lab struct{ l, a, b float64 }

func parseSRGB(hex string) (rgb, error) {
	if len(hex) != 7 || hex[0] != '#' {
		return rgb{}, fmt.Errorf("colour %q", hex)
	}
	var r, g, b uint8
	if _, err := fmt.Sscanf(hex, "#%02x%02x%02x", &r, &g, &b); err != nil {
		return rgb{}, fmt.Errorf("colour %q", hex)
	}
	return rgb{float64(r) / 255, float64(g) / 255, float64(b) / 255}, nil
}

func formatSRGB(colour rgb) string {
	return fmt.Sprintf("#%02X%02X%02X", channel(colour.r), channel(colour.g), channel(colour.b))
}

func channel(value float64) uint8 {
	if value < 0 {
		value = 0
	}
	if value > 1 {
		value = 1
	}
	return uint8(math.Round(value * 255))
}

// toOKLab turns a screen colour into L, a, and b.
// A stored channel is bent for the screen, so the first step undoes that bend
// and gets the real amount of light. Those amounts are then split the way the
// eye splits them, and the cube root stops a bright parent from crushing the mix.
func toOKLab(colour rgb) lab {
	red := linearise(colour.r)
	green := linearise(colour.g)
	blue := linearise(colour.b)

	long := math.Cbrt(0.4122214708*red + 0.5363325363*green + 0.0514459929*blue)
	medium := math.Cbrt(0.2119034982*red + 0.6806995451*green + 0.1073969566*blue)
	short := math.Cbrt(0.0883024619*red + 0.2817188376*green + 0.6299787005*blue)

	return lab{
		l: 0.2104542553*long + 0.7936177850*medium - 0.0040720468*short,
		a: 1.9779984951*long - 2.4285922050*medium + 0.4505937099*short,
		b: 0.0259040371*long + 0.7827717662*medium - 0.8086757660*short,
	}
}

// fromOKLab runs those steps backwards and bends the light for the screen again.
func fromOKLab(colour lab) rgb {
	long := colour.l + 0.3963377774*colour.a + 0.2158037573*colour.b
	medium := colour.l - 0.1055613458*colour.a - 0.0638541728*colour.b
	short := colour.l - 0.0894841775*colour.a - 1.2914855480*colour.b
	long = long * long * long
	medium = medium * medium * medium
	short = short * short * short

	return rgb{
		r: encode(4.0767416621*long - 3.3077115913*medium + 0.2309699292*short),
		g: encode(-1.2684380046*long + 2.6097574011*medium - 0.3413193965*short),
		b: encode(-0.0041960863*long - 0.7034186147*medium + 1.7076147010*short),
	}
}

func linearise(channel float64) float64 {
	if channel >= 0.04045 {
		return math.Pow((channel+0.055)/1.055, 2.4)
	}
	return channel / 12.92
}

func encode(light float64) float64 {
	if light >= 0.0031308 {
		return 1.055*math.Pow(light, 1/2.4) - 0.055
	}
	return 12.92 * light
}
