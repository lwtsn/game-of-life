import { useEffect, useRef, useState, type DragEvent, type PointerEvent } from 'react'
import type { Cell } from '../api/grid.ts'
import type { Pattern } from '../gen/life/v1/pattern_pb.js'
import { boardSize, cellAtPoint, cellNear, cellsAlong, COLS, fitCell, GAP, ROWS, stampOrigin, type PatternButton } from './patterns.ts'

const DEAD = '#DBE2EF'
const ALIVE = '#112D4E'
const GAP_COLOR = '#F9F7F7'

// Ghost is a shape being dragged over the board. It carries its cells so drawing needs no lookup.
type Ghost = { pattern: Pattern; x: number; y: number; cells: PatternButton['cells'] }

function draw(context: CanvasRenderingContext2D, cell: number, cells: Cell[] | null, ghost: Ghost | null, colour: string) {
  const { width, height } = boardSize(cell)
  context.fillStyle = GAP_COLOR
  context.fillRect(0, 0, width, height)

  const step = cell + GAP
  for (let y = 0; y < ROWS; y++) {
    for (let x = 0; x < COLS; x++) {
      const square = cells?.[y * COLS + x]
      context.fillStyle = square?.alive ? (square.colour ?? ALIVE) : DEAD
      context.fillRect(x * step, y * step, cell, cell)
    }
  }

  if (!ghost) return
  context.save()
  context.globalAlpha = 0.8
  context.fillStyle = colour
  context.strokeStyle = ALIVE
  context.lineWidth = 1
  for (const [dx, dy] of ghost.cells) {
    const x = ghost.x + dx
    const y = ghost.y + dy
    context.fillRect(x * step, y * step, cell, cell)
    context.strokeRect(x * step + 0.5, y * step + 0.5, cell - 1, cell - 1)
  }
  context.restore()
}

function placedIndexes(cells: Cell[] | null) {
  if (!cells) return []
  const marked: number[] = []
  cells.forEach((square, index) => {
    if (square.colour) marked.push(index)
  })
  return marked
}

type BoardProps = {
  cells: Cell[] | null
  patterns?: readonly PatternButton[]
  colour?: string
  dragPattern?: Pattern | null
  onPlace: (x: number, y: number) => void
  onStamp?: (pattern: Pattern, x: number, y: number) => void
}

const NO_PATTERNS: readonly PatternButton[] = []

export function Board({ cells, patterns = NO_PATTERNS, colour = ALIVE, dragPattern = null, onPlace, onStamp }: BoardProps) {
  const frameRef = useRef<HTMLDivElement>(null)
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const painting = useRef(false)
  const lastCell = useRef<{ col: number; row: number } | null>(null)
  const ghostKey = useRef('')
  const [cell, setCell] = useState(12)
  const [ghost, setGhost] = useState<Ghost | null>(null)

  useEffect(() => {
    const frame = frameRef.current
    if (!frame || typeof ResizeObserver === 'undefined') return

    const measure = () => {
      const style = getComputedStyle(frame)
      const padX = parseFloat(style.paddingLeft) + parseFloat(style.paddingRight)
      const padY = parseFloat(style.paddingTop) + parseFloat(style.paddingBottom)
      const width = frame.clientWidth - padX
      const height = frame.clientHeight - padY
      setCell(fitCell(width, height))
    }

    measure()
    const observer = new ResizeObserver(measure)
    observer.observe(frame)
    return () => observer.disconnect()
  }, [])

  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas) return
    const context = canvas.getContext('2d')
    if (!context) return

    const { width, height } = boardSize(cell)
    const ratio = window.devicePixelRatio || 1
    canvas.width = Math.round(width * ratio)
    canvas.height = Math.round(height * ratio)
    canvas.style.width = `${width}px`
    canvas.style.height = `${height}px`
    context.setTransform(ratio, 0, 0, ratio, 0, 0)
    draw(context, cell, cells, ghost, colour)
  }, [cell, cells, ghost, colour])

  function paintAt(clientX: number, clientY: number) {
    const canvas = canvasRef.current
    if (!canvas) return
    const rect = canvas.getBoundingClientRect()
    const index = cellAtPoint(clientX - rect.left, clientY - rect.top, cell)
    if (index === null) return
    const col = index % COLS
    const row = Math.floor(index / COLS)
    const from = lastCell.current
    const points = from ? cellsAlong(from.col, from.row, col, row) : [[col, row] as [number, number]]
    for (let i = from ? 1 : 0; i < points.length; i++) {
      const [x, y] = points[i]
      onPlace(x, y)
    }
    lastCell.current = { col, row }
  }

  function onPointerDown(event: PointerEvent<HTMLCanvasElement>) {
    if (event.button !== 0) return
    painting.current = true
    lastCell.current = null
    paintAt(event.clientX, event.clientY)
    if (typeof event.currentTarget.setPointerCapture === 'function') {
      event.currentTarget.setPointerCapture(event.pointerId)
    }
  }

  function onPointerMove(event: PointerEvent<HTMLCanvasElement>) {
    if (!painting.current) return
    paintAt(event.clientX, event.clientY)
  }

  function stopPainting() {
    painting.current = false
    lastCell.current = null
  }

  function originAt(clientX: number, clientY: number, pattern: Pattern): Ghost | null {
    const canvas = canvasRef.current
    const shape = patterns.find((item) => item.pattern === pattern)
    if (!canvas || !shape) return null
    const rect = canvas.getBoundingClientRect()
    const index = cellNear(clientX - rect.left, clientY - rect.top, cell)
    if (index === null) return null
    const origin = stampOrigin(index % COLS, Math.floor(index / COLS), shape.width, shape.height)
    return { pattern, x: origin.x, y: origin.y, cells: shape.cells }
  }

  function showGhost(clientX: number, clientY: number) {
    if (dragPattern === null) return
    const next = originAt(clientX, clientY, dragPattern)
    const key = next ? `${next.pattern}:${next.x}:${next.y}` : ''
    if (key === ghostKey.current) return
    ghostKey.current = key
    setGhost(next)
  }

  function clearGhost() {
    if (ghostKey.current === '') return
    ghostKey.current = ''
    setGhost(null)
  }

  function onDragOver(event: DragEvent<HTMLCanvasElement>) {
    event.preventDefault()
    showGhost(event.clientX, event.clientY)
  }

  function onDrop(event: DragEvent<HTMLCanvasElement>) {
    event.preventDefault()
    const carried = Number(event.dataTransfer.getData('text/plain'))
    const pattern = patterns.some((item) => item.pattern === carried) ? (carried as Pattern) : dragPattern
    clearGhost()
    if (pattern === null || !onStamp) return
    const origin = originAt(event.clientX, event.clientY, pattern)
    if (!origin) return
    onStamp(origin.pattern, origin.x, origin.y)
  }

  const marked = placedIndexes(cells)
  const label =
    marked.length === 0
      ? 'Game of Life board'
      : `Game of Life board, cells ${marked.join(' ')} highlighted`

  return (
    <div ref={frameRef} className="flex h-full min-h-0 w-full items-center justify-center p-4">
      <canvas
        ref={canvasRef}
        role="img"
        aria-label={label}
        className="cursor-pointer touch-none"
        onPointerDown={onPointerDown}
        onPointerMove={onPointerMove}
        onPointerUp={stopPainting}
        onPointerCancel={stopPainting}
        onDragOver={onDragOver}
        onDragLeave={clearGhost}
        onDrop={onDrop}
      />
    </div>
  )
}
