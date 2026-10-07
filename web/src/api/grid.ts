export const COLS = 80
export const ROWS = 50

export type GridFrame = {
  width: number
  height: number
  cells: number[]
}

export function readGridFrame(value: unknown): GridFrame | null {
  if (typeof value !== 'object' || value === null) return null

  const frame = value as Partial<GridFrame>
  if (frame.width !== COLS || frame.height !== ROWS) return null
  if (!Array.isArray(frame.cells) || frame.cells.length !== COLS * ROWS) {
    return null
  }

  return { width: COLS, height: ROWS, cells: frame.cells }
}
