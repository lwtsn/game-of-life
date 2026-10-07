# game_of_life

Monorepo. `web` is the React app. `server` is the Go service.

## Web

```sh
cd web
npm install
npm run dev
npm test
```

`npm test` is Vitest. Playwright and Chromium are installed for later latency and component tests. There is no integration suite on Playwright yet.

The board opens `ws://127.0.0.1:8080/ws` unless `VITE_WS_URL` is set. Start the Go server first. The page client lives in `web/src/api`. The hooks live in `web/src/hooks`.

## Server

```sh
cd server
go test ./...
go run ./cmd/server
```

The server listens on `127.0.0.1:8080` and pushes a full 80 by 50 grid to `/ws` once a second. `cmd/server` composes the Fx modules. In `internal/grid`, `frame.go` is one grid and `random.go` is the source that fills it. `api/hub.go` is the socket. Each `module.go` only wires Fx.

Regenerate mocks with `go tool mockery`.
