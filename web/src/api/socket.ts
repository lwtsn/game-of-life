import { sessionId } from './session.ts'

export function socketUrl() {
  const base = socketBase()
  if (base === '') return ''
  const join = base.includes('?') ? '&' : '?'
  return `${base}${join}session=${encodeURIComponent(sessionId())}`
}

function socketBase() {
  const configured = import.meta.env.VITE_WS_URL
  if (configured) return configured
  if (import.meta.env.MODE === 'test') return ''
  if (import.meta.env.DEV) return 'ws://127.0.0.1:8080/ws'
  const scheme = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  return `${scheme}//${window.location.host}/ws`
}
