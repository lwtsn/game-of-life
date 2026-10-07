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

The server listens on `127.0.0.1:8080`. Gin serves `GET /ws`, `POST /start`, and `POST /layout`. A client receives the current board on connect, then each update. The grid runs the simulation, one step a second, and `POST /start` calls that same start method. `POST /layout` takes a name and does not apply it yet. Connected users are keyed by IP. Reconnect and drop are stubs.

The grid stores the current board. The source reads that board when it produces the next values. `internal/grid/random` is that source for now. Replace it with the life source when that source holds the game state.

Regenerate mocks with `go tool mockery`.
