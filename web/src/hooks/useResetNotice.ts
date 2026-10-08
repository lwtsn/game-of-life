import { useEffect, useRef } from 'react'
import { readReset } from '../api/reset.ts'
import { announceReset } from '../components/ResetNotice.tsx'
import { useGameSocket } from './useGameSocket.ts'

export function useResetNotice() {
  const { lastJsonMessage } = useGameSocket()
  const seen = useRef<unknown>(undefined)

  useEffect(() => {
    if (Object.is(seen.current, lastJsonMessage)) return
    seen.current = lastJsonMessage
    const colour = readReset(lastJsonMessage)
    if (colour) announceReset(colour)
  }, [lastJsonMessage])
}
