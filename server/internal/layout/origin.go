package layout

import "math/rand/v2"

// Origin picks the top-left cell for a shape that fits on the board.
type Origin func(boardWidth, boardHeight, shapeWidth, shapeHeight int) (x, y int)

func randomOrigin(boardWidth, boardHeight, shapeWidth, shapeHeight int) (int, int) {
	spanX := boardWidth - shapeWidth + 1
	spanY := boardHeight - shapeHeight + 1
	if spanX < 1 || spanY < 1 {
		return 0, 0
	}
	return rand.IntN(spanX), rand.IntN(spanY)
}
