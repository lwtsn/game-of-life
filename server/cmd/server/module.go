package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"
	"game_of_life/server/api"
	"game_of_life/server/internal/grid"
	"game_of_life/server/internal/grid/life"
	"game_of_life/server/internal/user"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

type config struct {
	addr    string
	webRoot string
}

func module() fx.Option {
	return fx.Options(
		life.Module,
		grid.Module,
		user.Module,
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
	return config{addr: addr, webRoot: os.Getenv("WEB_ROOT")}
}

func provideHTTP(cfg config, handler api.Handler, rpc *connect.Server) *http.Server {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	handler.Register(engine)
	if cfg.webRoot != "" {
		mountWeb(engine, cfg.webRoot)
	}
	mux := http.NewServeMux()
	connecthttp.Mount(mux, rpc)
	mux.Handle("/", engine)
	return &http.Server{Addr: cfg.addr, Handler: corsLocal(mux)}
}

func start(lc fx.Lifecycle, server *http.Server, handler api.Handler, board grid.Grid) {
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
			board.Start(ctx)
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
