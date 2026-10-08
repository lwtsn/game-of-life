import { createClient } from '@connectrpc/connect'
import { createConnectTransport } from '@connectrpc/connect-web'
import { LayoutService, type Pattern } from '../gen/life/v1/pattern_pb.js'
import { sessionId } from './session.ts'

export function connectBaseUrl() {
  const configured = import.meta.env.VITE_WS_URL
  if (configured) return configured.replace(/^ws/, 'http').replace(/\/ws$/, '')
  if (import.meta.env.MODE === 'test') return ''
  return 'http://127.0.0.1:8080'
}

export async function placePattern(pattern: Pattern) {
  const baseUrl = connectBaseUrl()
  if (baseUrl === '') return
  const transport = createConnectTransport({ baseUrl })
  const client = createClient(LayoutService, transport)
  await client.place({ pattern }, { headers: { 'X-Session': sessionId() } })
}
