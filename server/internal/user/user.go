package user

import (
	"hash/fnv"
	"net"
)

// User is someone identified by an IP address.
type User interface {
	IP() string
	Colour() string
}

type user struct {
	ip     string
	colour string
}

func (u user) IP() string     { return u.ip }
func (u user) Colour() string { return u.colour }

// New builds the user for an address. IPv4-mapped IPv6 is the same user as the IPv4 form.
func New(host string) User {
	ip := canonicalIP(host)
	return user{ip: ip, colour: colourFor(ip)}
}

// palette starts with the style-guide navy, blue, and tint. The rest are
// variants spread around the wheel. Every pair is far enough apart to tell
// apart, and every entry still reads on the mist cells. A hundred addresses
// share this set. Three colours cannot.
var palette = []string{
	"#112D4E",
	"#3F72AF",
	"#7AA0CE",
	"#542F1C",
	"#8C6A21",
	"#664B0A",
	"#4D820D",
	"#1C5420",
	"#AE8F29",
	"#109E12",
	"#1C0A66",
	"#6836A1",
	"#4D0F62",
	"#6B245F",
	"#7C132D",
	"#28820D",
	"#2B8271",
	"#172697",
	"#47218C",
	"#AE2929",
	"#155B5A",
	"#0D8240",
	"#1C66BA",
	"#9E1078",
	"#541C3C",
	"#889717",
	"#107A9E",
	"#109E92",
	"#A13652",
	"#29AE4E",
	"#36A18F",
	"#293FAE",
	"#C313B4",
	"#0D4582",
	"#39620F",
	"#5CAE29",
	"#C37A13",
	"#1C9BBA",
	"#8C3E21",
	"#66150A",
	"#C34E13",
	"#9713C3",
	"#C3134B",
	"#5E6B24",
	"#94109E",
	"#AE2995",
	"#A13676",
	"#C31313",
	"#6B13C3",
	"#65109E",
	"#3F13C3",
	"#109E6E",
	"#2B5F82",
	"#1313C3",
	"#9E1046",
	"#C31377",
	"#9E1210",
	"#8D36A1",
	"#7B2B82",
	"#29AE29",
	"#A15436",
}

func colourFor(ip string) string {
	sum := fnv.New64a()
	_, _ = sum.Write([]byte(ip))
	return palette[sum.Sum64()%uint64(len(palette))]
}

func canonicalIP(host string) string {
	parsed := net.ParseIP(host)
	if parsed == nil {
		return host
	}
	if v4 := parsed.To4(); v4 != nil {
		return v4.String()
	}
	return parsed.String()
}
