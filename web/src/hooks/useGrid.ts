import { useEffect, useState } from 'react'
import { readGridFrame } from '../api/grid.ts'
import { useGameSocket } from './useGameSocket.ts'

export function useGrid() {
  const { lastJsonMessage } = useGameSocket()
  const [cells, setCells] = useState<number[] | null>(null)

  useEffect(() => {
    const frame = readGridFrame(lastJsonMessage)
    if (frame) setCells(frame.cells)
  }, [lastJsonMessage])

  return cells
}
