package websocket

import (
	"bytes"
	"context"
	"strings"
	"time"

	"game_of_life/server/internal/grid/mocks"
	"game_of_life/server/internal/grid/source"

	"github.com/coder/websocket"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("fan-out", func() {
	It("keeps serving other sockets when one stops reading", func() {
		small := fakeFrame{width: 1, height: 1, cells: []source.Cell{{Alive: true}}}
		board := mocks.NewMockGrid(GinkgoT())
		board.EXPECT().Current().Return(small).Times(2)
		srv, h := testServer(board)

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		DeferCleanup(cancel)
		dial := func(session string) *websocket.Conn {
			GinkgoHelper()
			conn, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(srv.URL, "http")+"/ws?session="+session, nil)
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(func() { conn.CloseNow() })
			return conn
		}

		// stuck never reads, so its TCP buffers fill and its writes start to block.
		dial("stuck")
		reader := dial("reader")
		reader.SetReadLimit(-1)
		readBoard(reader, ctx)

		big := fakeFrame{width: 400, height: 400, cells: make([]source.Cell, 400*400)}
		bigJSON, err := big.ToJson()
		Expect(err).NotTo(HaveOccurred())

		// Fewer boards than a mailbox holds, but far more bytes than the stuck socket's buffers.
		const boards = 48
		start := time.Now()
		for range boards {
			h.writeAll(bigJSON)
		}
		Expect(time.Since(start)).To(BeNumerically("<", 250*time.Millisecond))

		// Compare raw bytes. Decoding 480 KB of JSON per board under -race takes long enough
		// that the reader becomes a slow client itself and its writes time out.
		got := 0
		for got < boards {
			_, data, err := reader.Read(ctx)
			Expect(err).NotTo(HaveOccurred())
			if bytes.Equal(data, bigJSON) {
				got++
			}
		}

		// The stuck socket's write times out and it is dropped. The reader stays.
		Eventually(func() int { return len(h.conns()) }, 5*time.Second, 50*time.Millisecond).Should(Equal(1))
	})
})
