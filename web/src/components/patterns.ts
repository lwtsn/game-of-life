import { COLS, ROWS } from '../api/grid.ts'
import { Pattern } from '../gen/life/v1/pattern_pb.js'

export { COLS, ROWS }

export const GAP = 1

export const patternButtons = [
  { pattern: Pattern.BLOCK, label: 'Block' },
  { pattern: Pattern.BLINKER, label: 'Blinker' },
  { pattern: Pattern.GLIDER, label: 'Glider' },
  { pattern: Pattern.BEACON, label: 'Beacon' },
] as const

export function cellAtPoint(x: number, y: number, cell: number) {
  if (cell <= 0) return null
  const step = cell + GAP
  const col = Math.floor(x / step)
  const row = Math.floor(y / step)
  if (col < 0 || row < 0 || col >= COLS || row >= ROWS) return null
  const localX = x - col * step
  const localY = y - row * step
  if (localX < 0 || localY < 0 || localX >= cell || localY >= cell) return null
  return row * COLS + col
}

export function boardSize(cell: number) {
  return {
    width: COLS * cell + (COLS - 1) * GAP,
    height: ROWS * cell + (ROWS - 1) * GAP,
  }
}

// Largest integer cell that fits the content box. Falls back to 12 before layout.
export function fitCell(width: number, height: number) {
  if (width < 40 || height < 40) return 12
  const byWidth = Math.floor((width - GAP * (COLS - 1)) / COLS)
  const byHeight = Math.floor((height - GAP * (ROWS - 1)) / ROWS)
  return Math.max(2, Math.min(byWidth, byHeight))
}
