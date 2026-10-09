import { useCallback, useState } from 'react'
import useWebSocketImport from 'react-use-websocket'
import { socketUrl } from '../api/socket.ts'

type UseWebSocket = typeof useWebSocketImport

// react-use-websocket is CommonJS. Vite 8 hands the browser the module
// object, while Vitest hands over the function. Both are accepted here.
function resolveUseWebSocket(mod: UseWebSocket): UseWebSocket {
  if (typeof mod === 'function') {
    return mod
  }

  const withDefault = mod as unknown as { default?: UseWebSocket }
  if (typeof withDefault.default === 'function') {
    return withDefault.default
  }

  return mod
}

const useWebSocket = resolveUseWebSocket(useWebSocketImport)

// Connection is what the page can say about the socket.
// off: no socket configured (tests). lost: retries ran out and the player has to ask again.
export type Connection = 'off' | 'connecting' | 'open' | 'reconnecting' | 'lost'

// ReadyState.OPEN in react-use-websocket.
const OPEN = 1

// reconnectDelay doubles from one second up to thirty, with jitter so clients do not retry together.
export function reconnectDelay(attempt: number) {
  const backoff = Math.min(1_000 * 2 ** attempt, 30_000)
  const jitter = Math.random() * 500
  return backoff + jitter
}

export function useGameSocket() {
  const [attempt, setAttempt] = useState(0)
  const [gaveUp, setGaveUp] = useState(false)
  const [opened, setOpened] = useState(false)
  const base = socketUrl()
  // A new URL makes the library open a new socket. The server only reads session.
  const url = base.length > 0 && attempt > 0 ? `${base}&attempt=${attempt}` : base

  const socket = useWebSocket(
    url,
    {
      share: true,
      shouldReconnect: () => true,
      reconnectAttempts: 10,
      reconnectInterval: reconnectDelay,
      onOpen: () => {
        setOpened(true)
        setGaveUp(false)
      },
      onReconnectStop: () => setGaveUp(true),
    },
    url.length > 0,
  )

  let connection: Connection
  if (url.length === 0) connection = 'off'
  else if (gaveUp) connection = 'lost'
  else if (socket.readyState === OPEN) connection = 'open'
  else connection = opened ? 'reconnecting' : 'connecting'

  // Stable across renders, so it can sit in context without re-rendering consumers.
  const reconnect = useCallback(() => {
    setGaveUp(false)
    setAttempt((count) => count + 1)
  }, [])

  return { ...socket, connection, reconnect }
}
