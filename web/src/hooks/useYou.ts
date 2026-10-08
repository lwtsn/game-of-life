import { useEffect, useState } from 'react'
import { readYou } from '../api/you.ts'
import { useGameSocket } from './useGameSocket.ts'

export function useYou() {
  const { lastJsonMessage } = useGameSocket()
  const [you, setYou] = useState<string | null>(null)

  useEffect(() => {
    const colour = readYou(lastJsonMessage)
    if (colour) setYou(colour)
  }, [lastJsonMessage])

  return you
}
