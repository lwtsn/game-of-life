package websocket

import (
	"net/url"
	"path"
	"strings"
)

// Origins lists the host patterns that may open the socket or call Connect from another origin.
// Patterns use path.Match syntax against host:port, for example "game.example.com" or "*.example.com:*".
// The socket always accepts a page served from its own host.
type Origins []string

// DefaultOrigins allows pages on this machine, such as the Vite dev server.
var DefaultOrigins = Origins{"localhost", "localhost:*", "127.0.0.1", "127.0.0.1:*"}

// ParseOrigins reads a comma-separated list. An empty list gives DefaultOrigins.
func ParseOrigins(raw string) Origins {
	var out Origins
	for _, part := range strings.Split(raw, ",") {
		if pattern := strings.TrimSpace(part); pattern != "" {
			out = append(out, pattern)
		}
	}
	if len(out) == 0 {
		return DefaultOrigins
	}
	return out
}

// Patterns returns the configured patterns, or DefaultOrigins when none are set.
func (o Origins) Patterns() []string {
	if len(o) == 0 {
		return DefaultOrigins
	}
	return o
}

// Allows reports whether an Origin header matches a pattern, the same way coder/websocket matches OriginPatterns.
func (o Origins) Allows(origin string) bool {
	parsed, err := url.Parse(origin)
	if err != nil || parsed.Host == "" {
		return false
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return false
	}
	host := strings.ToLower(parsed.Host)
	for _, pattern := range o.Patterns() {
		if matched, _ := path.Match(strings.ToLower(pattern), host); matched {
			return true
		}
	}
	return false
}
