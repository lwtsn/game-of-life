import { useEffect, useRef, useState } from 'react'
import { boardSize, COLS, fitCell, GAP, ROWS } from './patterns.ts'

const DEAD = '#DBE2EF'
const ALIVE = '#112D4E'
const GAP_COLOR = '#F9F7F7'

function draw(
  context: CanvasRenderingContext2D,
  cell: number,
  cells: number[] | null,
) {
  const { width, height } = boardSize(cell)
  context.fillStyle = GAP_COLOR
  context.fillRect(0, 0, width, height)

  const step = cell + GAP
  for (let y = 0; y < ROWS; y++) {
    for (let x = 0; x < COLS; x++) {
      const alive = cells !== null && cells[y * COLS + x] === 1
      context.fillStyle = alive ? ALIVE : DEAD
      context.fillRect(x * step, y * step, cell, cell)
    }
  }
}

type BoardProps = {
  cells: number[] | null
}

export function Board({ cells }: BoardProps) {
  const frameRef = useRef<HTMLDivElement>(null)
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const [cell, setCell] = useState(12)

  useEffect(() => {
    const frame = frameRef.current
    if (!frame || typeof ResizeObserver === 'undefined') return

    const measure = () => {
      const style = getComputedStyle(frame)
      const padX =
        parseFloat(style.paddingLeft) + parseFloat(style.paddingRight)
      const padY =
        parseFloat(style.paddingTop) + parseFloat(style.paddingBottom)
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
    draw(context, cell, cells)
  }, [cell, cells])

  return (
    <div
      ref={frameRef}
      className="flex h-svh w-full items-center justify-center px-16 py-8"
    >
      <canvas ref={canvasRef} role="img" aria-label="Game of Life board" />
    </div>
  )
}
