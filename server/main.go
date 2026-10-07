package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"

	"game_of_life/server/api"
	"game_of_life/server/internal/grid"
	"go.uber.org/fx"
)

type Config struct {
	Addr string
}

func main() {
	fx.New(module()).Run()
}

func module() fx.Option {
	return fx.Options(
		fx.Provide(loadConfig, newSource, api.NewHub, newMux, newHTTPServer),
		fx.Invoke(start),
	)
}

func loadConfig() Config {
	addr := os.Getenv("GAME_ADDR")
	if addr == "" {
		addr = "127.0.0.1:8080"
	}
	return Config{Addr: addr}
}

func newSource() grid.Source {
	return grid.NewRandom()
}

func newMux(hub *api.Hub) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("GET /ws", hub)
	return mux
}

func newHTTPServer(cfg Config, mux *http.ServeMux) *http.Server {
	return &http.Server{Addr: cfg.Addr, Handler: mux}
}

func start(lc fx.Lifecycle, server *http.Server, hub *api.Hub) {
	var cancel context.CancelFunc
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			ln, err := net.Listen("tcp", server.Addr)
			if err != nil {
				return err
			}
			log.Printf("websocket listening on ws://%s/ws", ln.Addr())
			var ctx context.Context
			ctx, cancel = context.WithCancel(context.Background())
			go hub.Run(ctx)
			go func() {
				err := server.Serve(ln)
				if err != nil && !errors.Is(err, http.ErrServerClosed) {
					log.Printf("http server: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			if cancel != nil {
				cancel()
			}
			return server.Shutdown(ctx)
		},
	})
}
