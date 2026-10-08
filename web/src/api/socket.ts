import { sessionId } from './session.ts'

export function socketUrl() {
  const configured = import.meta.env.VITE_WS_URL
  const base = configured || (import.meta.env.MODE === 'test' ? '' : 'ws://127.0.0.1:8080/ws')
  if (base === '') return ''
  const join = base.includes('?') ? '&' : '?'
  return `${base}${join}session=${encodeURIComponent(sessionId())}`
}
