import { useEffect, useState } from 'react'
import { readGridFrame, type Cell } from '../api/grid.ts'
import { useGameSocket } from './useGameSocket.ts'

export function useGrid() {
  const { lastJsonMessage, sendJsonMessage } = useGameSocket()
  const [cells, setCells] = useState<Cell[] | null>(null)

  useEffect(() => {
    const frame = readGridFrame(lastJsonMessage)
    if (frame) setCells(frame.cells)
  }, [lastJsonMessage])

  function place(x: number, y: number) {
    sendJsonMessage({ x, y })
  }

  function reset() {
    sendJsonMessage({ reset: true })
  }

  return { cells, place, reset }
}
