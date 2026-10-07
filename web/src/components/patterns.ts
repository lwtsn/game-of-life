export const COLS = 80
export const ROWS = 50
export const GAP = 1

export const patterns = [
  'Glider',
  'Blinker',
  'Toad',
  'Beacon',
  'Pulsar',
  'Lightweight spaceship',
  'Gosper glider gun',
  'Block',
  'Beehive',
  'Loaf',
] as const

export type PatternName = (typeof patterns)[number]

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
