import { useEffect, useSyncExternalStore } from 'react'
import { announcePresence } from '../components/PresenceNotice.tsx'
import { useGameClient } from './context.ts'

// Each hook reads one slice of the store. A component re-renders only when its slice changes.

export function useBoard() {
  const { store, commands } = useGameClient()
  const { cells } = useSyncExternalStore(store.subscribe, store.getBoard)
  return { cells, place: commands.place, reset: commands.reset }
}

export function useFrame() {
  const { store } = useGameClient()
  return useSyncExternalStore(store.subscribe, store.getFrame)
}

export function useClock() {
  const { store, commands } = useGameClient()
  const clock = useSyncExternalStore(store.subscribe, store.getClock)
  return { ...clock, setRunning: commands.setRunning, setPace: commands.setPace }
}

export function usePeople() {
  const { store, commands } = useGameClient()
  const people = useSyncExternalStore(store.subscribe, store.getPeople)
  return { ...people, chooseColour: commands.chooseColour }
}

export function useConnection() {
  const { connection, reconnect } = useGameClient()
  return { connection, reconnect }
}

// useCommands gives the actions without subscribing to any state.
export function useCommands() {
  return useGameClient().commands
}

// usePresenceToasts shows a toast when someone joins, leaves, resets or changes colour. Mount it once.
export function usePresenceToasts() {
  const { store } = useGameClient()
  useEffect(() => store.onNotice(({ colour, text }) => announcePresence(colour, text)), [store])
}
