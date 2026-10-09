import { fromJson } from '@bufbuild/protobuf'
import { COLS, ROWS } from '../api/grid.ts'
import { CatalogueSchema, type Pattern, type Shape } from '../gen/life/v1/pattern_pb.js'
import catalogueJson from '../gen/life/v1/catalogue.json' with { type: 'json' }

export { COLS, ROWS }

export const GAP = 1

export type PatternButton = {
  pattern: Pattern
  label: string
  width: number
  height: number
  cells: Array<[number, number]>
}

// catalogue.json is protojson of life.v1.Catalogue, exported from patterns.textproto.
const catalogue = fromJson(CatalogueSchema, catalogueJson)

function buttonFrom(shape: Shape): PatternButton {
  let maxX = 0
  let maxY = 0
  const cells: Array<[number, number]> = []
  for (const cell of shape.cells) {
    const x = cell.x
    const y = cell.y
    if (x > maxX) maxX = x
    if (y > maxY) maxY = y
    cells.push([x, y])
  }
  return {
    pattern: shape.pattern,
    label: shape.label,
    width: maxX + 1,
    height: maxY + 1,
    cells,
  }
}

export const patternButtons: PatternButton[] = catalogue.shapes.map(buttonFrom)

export function stampOrigin(col: number, row: number, width: number, height: number) {
  return {
    x: Math.max(0, Math.min(col, COLS - width)),
    y: Math.max(0, Math.min(row, ROWS - height)),
  }
}

export function cellNear(x: number, y: number, cell: number) {
  if (cell <= 0) return null
  const step = cell + GAP
  const col = Math.min(COLS - 1, Math.max(0, Math.floor(x / step)))
  const row = Math.min(ROWS - 1, Math.max(0, Math.floor(y / step)))
  return row * COLS + col
}

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

// Cells a stroke crosses, including both ends. Samples from a fast drag are joined so none are skipped.
export function cellsAlong(fromCol: number, fromRow: number, toCol: number, toRow: number) {
  const points: Array<[number, number]> = []
  const dx = Math.abs(toCol - fromCol)
  const dy = Math.abs(toRow - fromRow)
  const sx = fromCol < toCol ? 1 : -1
  const sy = fromRow < toRow ? 1 : -1
  let err = dx - dy
  let x = fromCol
  let y = fromRow
  while (true) {
    points.push([x, y])
    if (x === toCol && y === toRow) return points
    const doubled = 2 * err
    if (doubled > -dy) {
      err -= dy
      x += sx
    }
    if (doubled < dx) {
      err += dx
      y += sy
    }
  }
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
