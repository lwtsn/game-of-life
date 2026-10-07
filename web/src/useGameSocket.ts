import useWebSocketImport, { ReadyState } from 'react-use-websocket'

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

function defaultSocketUrl() {
  const configured = import.meta.env.VITE_WS_URL
  if (configured) return configured
  if (import.meta.env.MODE === 'test') return ''
  return 'ws://127.0.0.1:8080/ws'
}

export function useGameSocket() {
  const socketUrl = defaultSocketUrl()

  const socket = useWebSocket(
    socketUrl,
    {
      share: true,
      shouldReconnect: () => true,
      reconnectAttempts: 10,
      reconnectInterval: 3_000,
    },
    socketUrl.length > 0,
  )

  return { socketUrl, ...socket }
}

export { ReadyState }
