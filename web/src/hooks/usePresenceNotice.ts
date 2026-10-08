import { useEffect, useRef } from 'react'
import { readArrival } from '../api/arrival.ts'
import { announcePresence } from '../components/PresenceNotice.tsx'
import { useGameSocket } from './useGameSocket.ts'

export function usePresenceNotice() {
  const { lastJsonMessage } = useGameSocket()
  const seen = useRef<unknown>(undefined)

  useEffect(() => {
    if (Object.is(seen.current, lastJsonMessage)) return
    seen.current = lastJsonMessage
    const notice = readArrival(lastJsonMessage)
    if (notice) announcePresence(notice.colour, notice.text)
  }, [lastJsonMessage])
}
