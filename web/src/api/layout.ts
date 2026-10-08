import { createClient } from '@connectrpc/connect'
import { createConnectTransport } from '@connectrpc/connect-web'
import { LayoutService, type Pattern } from '../gen/life/v1/pattern_pb.js'
import { sessionId } from './session.ts'

function connectBaseUrl() {
  const configured = import.meta.env.VITE_WS_URL
  if (configured) return configured.replace(/^ws/, 'http').replace(/\/ws$/, '')
  if (import.meta.env.MODE === 'test') return ''
  if (import.meta.env.DEV) return 'http://127.0.0.1:8080'
  return window.location.origin
}

export async function placePattern(pattern: Pattern, x: number, y: number) {
  const baseUrl = connectBaseUrl()
  if (baseUrl === '') return
  const transport = createConnectTransport({ baseUrl })
  const client = createClient(LayoutService, transport)
  await client.place({ pattern, origin: { x, y } }, { headers: { 'X-Session': sessionId() } })
}
