export const COLS = 80
export const ROWS = 50

const hex = /^#[0-9A-Fa-f]{6}$/

export type Cell = {
  alive: boolean
  id?: string
  colour?: string
}

export type GridFrame = {
  width: number
  height: number
  cells: Cell[]
}

function readCell(value: unknown): Cell | null {
  if (typeof value !== 'object' || value === null || Array.isArray(value)) return null

  const raw = value as Record<string, unknown>
  if (typeof raw.alive !== 'boolean') return null

  const cell: Cell = { alive: raw.alive }
  if ('id' in raw) {
    if (typeof raw.id !== 'string' || raw.id.length === 0) return null
    cell.id = raw.id
  }
  if ('colour' in raw) {
    if (typeof raw.colour !== 'string' || !hex.test(raw.colour)) return null
    cell.colour = raw.colour
  }
  return cell
}

export function readGridFrame(value: unknown): GridFrame | null {
  if (typeof value !== 'object' || value === null) return null

  const frame = value as { width?: unknown; height?: unknown; cells?: unknown }
  if (frame.width !== COLS || frame.height !== ROWS) return null
  if (!Array.isArray(frame.cells) || frame.cells.length !== COLS * ROWS) return null

  const cells: Cell[] = []
  for (const item of frame.cells) {
    const cell = readCell(item)
    if (!cell) return null
    cells.push(cell)
  }

  return { width: COLS, height: ROWS, cells }
}
