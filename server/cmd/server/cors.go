package main

import (
	"net/http"
	"net/url"
)

func corsLocal(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if originAllowed(r.Header.Get("Origin")) {
			header := w.Header()
			header.Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
			header.Add("Vary", "Origin")
			header.Set("Access-Control-Allow-Methods", "POST, OPTIONS")
			header.Set("Access-Control-Allow-Headers", "Content-Type, Connect-Protocol-Version, Connect-Timeout-Ms, Connect-Accept-Encoding, Connect-Content-Encoding, Grpc-Timeout, X-Grpc-Web, X-User-Agent")
			header.Set("Access-Control-Expose-Headers", "Connect-Content-Encoding")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func originAllowed(origin string) bool {
	parsed, err := url.Parse(origin)
	if err != nil || parsed.Host == "" {
		return false
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return false
	}
	switch parsed.Hostname() {
	case "localhost", "127.0.0.1":
		return true
	default:
		return false
	}
}
