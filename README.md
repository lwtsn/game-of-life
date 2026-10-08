# Game of Life

## What is it
This is an implementation of Conway's Game of Life. 

The game utilises an 80x50 grid made up of individual blocks or cells. 

Every tick (default is one second) the game runs a simulation. The simulation rules are as follows:
For each cell on the grid:
If it has less than 2 neighbours (including diagonal) it will be removed (died due to under-population)
If it has 2 or 3 neighbours it will survive
If it has over 3 neighbours it will be removed (died due to over-population)
If an empty (or dead) cell has 3 adjacent neighbours it will be born


## Design decisions

### Heavy computational logic is done in the backend in Golang.
The frontend (React & HTML canvas) are used purely for display and interaction

Trade offs:
- The project requires both a backend and frontend server to run to work.
- Deployment time is slightly slower (albeit still fast) due to the Go dependency.

### The board is not infinite
The board is a fixed 80x50 grid. This was decided both for visual clarity and speed of development.

Trade offs:
- Towards each edge of the board the program has some unexpected behaviours as I consider anything outside the scope of the board to a dead cell. This leads to cells resting in a fixed state when touching the edge
- Users are limited in what they can build

### Protobuf is the contract between the page and the server
Cross service communication lines are build using `.proto` files. This is to ensure at compile time that Go and Typescript cannot drift apart.

Trade offs:
- There is a generate step (`buf generate`) whenever the contract changes
- Slightly more ceremony than plain JSON for a small project, but worth it so the frontend and backend stay aligned

### There are two communication channels
The grid receives updates via websockets and API calls are made through RPC-Connect.

Trade offs:
- Two ways of talking to the server instead of one
- RPC-Connect makes non-standard API calls which may throw off frontend engineers
- As above `buf generate` is required to update API calls. 

### Uber FX is used for DI. Interfaces are exported not concrete types

Trade offs:
- More up-front wiring
- Bubble up modules can be harder to debug in complex applications (mitigated with isolated integration tests) 
- Interface collision can cause run-time panics

### Birth colour is mixed in OKLab
Rather than create a generic average of the parent colours I use OKLab to retain brightness and feel of the colour giving a more natural looking transition 

Trade offs:
- OKLab is less obvious than a simple RGB average if you have not seen it before
- Extra conversion on each birth (cheap at this board size)

### Stop drops anything still queued
Once stop is clicked I drop any updated which are queued 

Trade offs:
- A frame that was mid-flight can be discarded even though the server had already computed it
- Stop feels immediate, which is what you want when playing

### Hashlife was not used
The algorithm I used was profiled to ~10k updates per second on a 80x50 board (on an M1 mac)
I also profiled UI updates in the page and by hand. 
The applied-updates counter could reach about 196/s on an M1; above 100/s the experience was already poor so an optimised algorithm would not solve our bottleneck.

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
