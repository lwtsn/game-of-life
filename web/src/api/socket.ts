export function socketUrl() {
  const configured = import.meta.env.VITE_WS_URL
  if (configured) return configured
  if (import.meta.env.MODE === 'test') return ''
  return 'ws://127.0.0.1:8080/ws'
}
