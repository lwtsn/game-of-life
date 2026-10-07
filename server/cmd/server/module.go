package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"

	"game_of_life/server/api"
	"game_of_life/server/internal/grid/random"
	"go.uber.org/fx"
)

type config struct {
	addr string
}

func module() fx.Option {
	return fx.Options(
		random.Module,
		api.Module,
		fx.Provide(provideConfig, provideHTTP),
		fx.Invoke(start),
	)
}

func provideConfig() config {
	addr := os.Getenv("GAME_ADDR")
	if addr == "" {
		addr = "127.0.0.1:8080"
	}
	return config{addr: addr}
}

func provideHTTP(cfg config, handler api.Handler) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("GET /ws", handler)
	return &http.Server{Addr: cfg.addr, Handler: mux}
}

func start(lc fx.Lifecycle, server *http.Server, handler api.Handler) {
	var cancel context.CancelFunc
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			ln, err := net.Listen("tcp", server.Addr)
			if err != nil {
				return err
			}
			// Addr stays ":0" until we keep the port Listen chose.
			server.Addr = ln.Addr().String()
			log.Printf("websocket listening on ws://%s/ws", server.Addr)
			var ctx context.Context
			ctx, cancel = context.WithCancel(context.Background())
			go handler.Run(ctx)
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
