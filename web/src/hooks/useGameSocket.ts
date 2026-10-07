import useWebSocketImport, { ReadyState } from 'react-use-websocket'
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

export function useGameSocket() {
  const url = socketUrl()

  const socket = useWebSocket(
    url,
    {
      share: true,
      shouldReconnect: () => true,
      reconnectAttempts: 10,
      reconnectInterval: 3_000,
    },
    url.length > 0,
  )

  return { socketUrl: url, ...socket }
}

export { ReadyState }
