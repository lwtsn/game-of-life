import { useEffect, useState } from 'react'
import { readColours } from '../api/colours.ts'
import { useGameSocket } from './useGameSocket.ts'

export function usePresence() {
  const { lastJsonMessage } = useGameSocket()
  const [colours, setColours] = useState<string[]>([])

  useEffect(() => {
    const next = readColours(lastJsonMessage)
    if (next) setColours(next)
  }, [lastJsonMessage])

  return colours
}
