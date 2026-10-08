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

`go test ./...` runs the Ginkgo suites. The assertions are Gomega.

The server listens on `127.0.0.1:8080`. `server/api/router.go` mounts the routes. `websocket` serves `GET /ws` and keeps every connected socket. An address still has one user and one colour. `controls` serves `POST /start` and `POST /layout`. `internal/user` keeps one user per address, and that user carries the colour. The service joins and leaves that set and returns the colours. The set is the navy, blue, and tint from the style guide, plus variants spread around the wheel. The socket sends the connected colours as `{"colours":["#RRGGBB"]}`. A client receives the current board on connect, then each update. The grid runs the simulation, one step a second, and `POST /start` calls that same start method. `POST /layout` takes a name and does not apply it yet. Reconnect and drop are stubs.

The grid stores the current board. The source reads that board when it produces the next values. `internal/grid/life` is that source. The first board is an empty 80 by 50.

`internal/grid/life` on an Apple M1, three runs:

```sh
cd server && go test -run='^$' -bench=BenchmarkNext -benchmem ./internal/grid/life
```

| Board | Time per update | Updates per second | Memory |
|---|---|---|---|
| 80×50 | 102 µs | about 9,800 | 96 KB, 3 allocs |
| 1000×1000 | 24.6 ms | about 41 | 23 MB, 3 allocs |

Regenerate mocks with `go tool mockery`.
