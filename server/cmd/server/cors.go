package main

import (
	"net/http"

	"game_of_life/server/api/websocket"
)

func withCORS(next http.Handler, origins websocket.Origins) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origins.Allows(r.Header.Get("Origin")) {
			header := w.Header()
			header.Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
			header.Add("Vary", "Origin")
			header.Set("Access-Control-Allow-Methods", "POST, OPTIONS")
			header.Set("Access-Control-Allow-Headers", "Content-Type, Connect-Protocol-Version, Connect-Timeout-Ms, Connect-Accept-Encoding, Connect-Content-Encoding, Grpc-Timeout, X-Grpc-Web, X-User-Agent, X-Session")
			header.Set("Access-Control-Expose-Headers", "Connect-Content-Encoding")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
