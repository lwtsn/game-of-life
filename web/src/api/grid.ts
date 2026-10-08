import { fromJson, type JsonValue } from '@bufbuild/protobuf'
import { GridSize, ServerMessageSchema, type Cell as ProtoCell } from '../gen/life/v1/socket_pb.js'
import { isHex } from './hex.ts'

export const COLS = GridSize.WIDTH
export const ROWS = GridSize.HEIGHT

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

function readCell(cell: ProtoCell): Cell | null {
  if (cell.colour !== '' && !isHex(cell.colour)) return null

  const read: Cell = { alive: cell.alive }
  if (cell.id !== '') read.id = cell.id
  if (cell.colour !== '') read.colour = cell.colour
  return read
}

export function readGridFrame(value: unknown): GridFrame | null {
  if (typeof value !== 'object' || value === null || Array.isArray(value)) return null

  let message
  try {
    message = fromJson(ServerMessageSchema, value as JsonValue)
  } catch {
    return null
  }
  if (message.width !== COLS || message.height !== ROWS) return null
  if (message.cells.length !== COLS * ROWS) return null

  const cells: Cell[] = []
  for (const item of message.cells) {
    const cell = readCell(item)
    if (!cell) return null
    cells.push(cell)
  }

  return { width: COLS, height: ROWS, cells }
}
