import { createClient, type Client } from '@connectrpc/connect'
import { createConnectTransport } from '@connectrpc/connect-web'
import { LayoutService, type Pattern, type Shape } from '../gen/life/v1/pattern_pb.js'
import { sessionId } from './session.ts'

function connectBaseUrl() {
  const configured = import.meta.env.VITE_WS_URL
  if (configured) return configured.replace(/^ws/, 'http').replace(/\/ws$/, '')
  if (import.meta.env.MODE === 'test') return ''
  if (import.meta.env.DEV) return 'http://127.0.0.1:8080'
  return window.location.origin
}

// One client for the page, made on first use. It is null in unit tests, where there is no server.
let layout: Client<typeof LayoutService> | null | undefined

function layoutClient() {
  if (layout === undefined) {
    const baseUrl = connectBaseUrl()
    layout = baseUrl === '' ? null : createClient(LayoutService, createConnectTransport({ baseUrl }))
  }
  return layout
}

// listPatterns asks the server which shapes it can stamp.
export async function listPatterns(): Promise<Shape[]> {
  const client = layoutClient()
  if (!client) return []
  const { shapes } = await client.listPatterns({})
  return shapes
}

export async function placePattern(pattern: Pattern, x: number, y: number) {
  const client = layoutClient()
  if (!client) return
  await client.place({ pattern, origin: { x, y } }, { headers: { 'X-Session': sessionId() } })
}
