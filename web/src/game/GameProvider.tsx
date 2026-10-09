import { useEffect, useMemo, useRef, useState, type ReactNode } from 'react'
import type { JsonValue } from '@bufbuild/protobuf'
import { readSocketMessage } from '../api/socketMessage.ts'
import { useGameSocket } from '../hooks/useGameSocket.ts'
import { createCommands } from './commands.ts'
import { GameContext, type GameClient } from './context.ts'
import { GameStore } from './store.ts'

// GameProvider owns the one socket and the one store. Everything below it reads slices through hooks,
// so a new board re-renders the board and not the rest of the page.
export function GameProvider({ children }: { children: ReactNode }) {
  const [store] = useState(() => new GameStore())
  const { lastJsonMessage, sendJsonMessage, connection, reconnect } = useGameSocket()

  // StrictMode runs effects twice in development. The same message must not be applied twice.
  const seen = useRef<unknown>(undefined)
  useEffect(() => {
    if (Object.is(seen.current, lastJsonMessage)) return
    seen.current = lastJsonMessage
    const message = readSocketMessage(lastJsonMessage)
    if (message) store.apply(message)
  }, [lastJsonMessage, store])

  useEffect(() => {
    const id = window.setInterval(() => store.tick(), 250)
    return () => window.clearInterval(id)
  }, [store])

  // sendJsonMessage is stable in react-use-websocket, so commands are one object for the life of the page
  // and passing them down never re-renders anything.
  const commands = useMemo(
    () => createCommands((message: JsonValue) => sendJsonMessage(message), store),
    [sendJsonMessage, store],
  )

  const client = useMemo<GameClient>(
    () => ({ store, commands, connection, reconnect }),
    [store, commands, connection, reconnect],
  )

  return <GameContext value={client}>{children}</GameContext>
}
