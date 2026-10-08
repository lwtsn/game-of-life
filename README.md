# Game of Life

## What is it
This is an implementation of Conway's Game of Life.

The game uses an 80x50 grid made up of individual blocks or cells.

Every tick (default is one second) the game runs a simulation. The rules are as follows.
For each cell on the grid:
- If it has fewer than 2 neighbours (including diagonals) it is removed (under-population)
- If it has 2 or 3 neighbours it survives
- If it has more than 3 neighbours it is removed (over-population)
- If an empty (dead) cell has exactly 3 neighbours it is born

## Design decisions

### Heavy computational logic is done in the backend in Golang.
The frontend (React and HTML canvas) is used purely for display and interaction.

Trade offs:
- The project needs both a backend and a frontend running to work
- Deployment is slightly slower (still fast) because of the Go build

### The board is not infinite
The board is a fixed 80x50 grid. This was decided for visual clarity and speed of development.

Trade offs:
- Towards each edge the behaviour is odd, because anything outside the board counts as a dead cell, so patterns that touch the rim can settle into a fixed state
- Users are limited in what they can build

### Protobuf is the contract between the page and the server
Cross-service communication is built from `.proto` files so that at compile time Go and TypeScript cannot drift apart.

Trade offs:
- There is a generate step (`buf generate`) whenever the contract changes
- Slightly more ceremony than plain JSON for a small project, but worth it so the frontend and backend stay aligned

### There are two communication channels
The grid receives updates over websockets. API calls go through Connect-RPC.

Trade offs:
- Two ways of talking to the server instead of one
- Connect-RPC is less familiar than plain REST and can surprise frontend engineers
- As above, `buf generate` is required when those API calls change

### Uber FX is used for DI. Interfaces are exported not concrete types
Only interfaces cross package boundaries. Concrete types stay inside each package, and FX wires the process.

Trade offs:
- More up-front wiring
- Bubble-up modules can be harder to debug in complex applications (mitigated with isolated integration tests)
- Interface collisions can cause run-time panics

### Birth colour is mixed in OKLab
Rather than a plain average of the parent colours, I use OKLab so the mix keeps the brightness and feel of the parents and looks more natural.

Trade offs:
- OKLab is less obvious than a simple RGB average if you have not seen it before
- Extra conversion on each birth (cheap at this board size)

### Stop drops anything still queued
Once stop is clicked all queued updates are dropped.

Trade offs:
- A frame that was mid-flight can be discarded even though the server had already computed it
- Stop feels immediate, which is what you want when playing

### Hashlife was not used
The algorithm I used was profiled to about 10k updates per second on an 80x50 board (on an M1 Mac).
I also profiled UI updates in the page and by hand.
The applied-updates counter could reach about 196/s on an M1; above 100/s the experience was already poor, so a faster algorithm would not fix the bottleneck.

Trade offs:
- The backend runs a less efficient algorithm
- The step stays simple enough to walk through and to interrupt safely

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

Each socket receives the whole board on connect and when the board changes. The clock starts at one generation a second. Any session can change that pace, from one to one hundred a second, or stop it. A new live cell takes the average colour of the three neighbours that produced it, and the toolbar places Block, Blinker, Glider, or Beacon through Connect `Place`.

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
