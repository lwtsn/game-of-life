import { createContext, useContext } from 'react'
import type { Connection } from '../hooks/useGameSocket.ts'
import type { GameCommands } from './commands.ts'
import type { GameStore } from './store.ts'

export type GameClient = {
  readonly store: GameStore
  readonly commands: GameCommands
  readonly connection: Connection
  readonly reconnect: () => void
}

export const GameContext = createContext<GameClient | null>(null)

export function useGameClient(): GameClient {
  const client = useContext(GameContext)
  if (!client) throw new Error('Game hooks must be used inside <GameProvider>.')
  return client
}
