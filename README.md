# game_of_life

This repository is the Game of Life exercise. The page is a React app in `web`, and the simulation is a Go service in `server`.

## On your machine

Go 1.26 and Node 22 are enough. Start the server before the page.

```sh
cd server
go test ./...
go run ./cmd/server
```

That listens on `127.0.0.1:8080`. `go test ./...` runs the Ginkgo suites, and the assertions are Gomega.

```sh
cd web
npm install
npm test
npm run dev
```

Open http://localhost:5173. The board uses `ws://127.0.0.1:8080/ws` unless `VITE_WS_URL` is set. `npm test` is Vitest.

## With Docker

From the repository root, build an image that contains the page and the server, then run it.

```sh
docker build -t game-of-life .
docker run --rm -p 8080:8080 game-of-life
```

Open http://localhost:8080. The same process serves the built files and the socket. Inside the container, `GAME_ADDR` is `0.0.0.0:8080` and `WEB_ROOT` is `/web`. Leave `WEB_ROOT` unset when you use `go run`, or the process will also try to serve a built page.

## What the server keeps

The page stores a session id in localStorage and sends it on the socket, and the server keeps one colour for that id. A refresh uses the same id, so the colour stays. A private window is a second person. Two tabs in the same window share the id and stay one person.

Each socket receives the whole board on connect and again after every change. The grid advances one step a second, a new live cell takes the average colour of the three neighbours that produced it, and the toolbar places Block, Blinker, Glider, or Beacon through Connect `Place`.

## Bench

`internal/grid/life` on an Apple M1, three runs:

```sh
cd server && go test -run='^$' -bench=BenchmarkNext -benchmem ./internal/grid/life
```

| Board | Time per update | Updates per second | Memory |
|---|---|---|---|
| 80×50 | 102 µs | about 9,800 | 96 KB, 3 allocs |
| 1000×1000 | 24.6 ms | about 41 | 23 MB, 3 allocs |

Regenerate mocks from `server/` with `go tool mockery`.
