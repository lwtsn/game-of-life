import { describe, expect, it, vi } from 'vitest'
import { MessageType } from '../gen/life/v1/socket_pb.js'
import { createCommands } from './commands.ts'
import { GameStore } from './store.ts'

function setup() {
  const store = new GameStore(() => 0)
  const send = vi.fn()
  return { store, send, commands: createCommands(send, store) }
}

describe('createCommands', () => {
  it('sends a placed cell', () => {
    const { send, commands } = setup()
    commands.place(3, 4)
    expect(send).toHaveBeenCalledWith({ point: { x: 3, y: 4 } })
  })

  it('stops locally before asking the server to stop', () => {
    const { store, send, commands } = setup()
    commands.setRunning(false)
    expect(store.getClock().running).toBe(false)
    expect(send).toHaveBeenCalledWith({ playback: 'PLAYBACK_STOPPED' })
  })

  it('clears the board on reset and ignores the boards still on their way', () => {
    const { store, send, commands } = setup()
    store.apply({ type: MessageType.BOARD, cells: [{ alive: true }], frame: 1 })
    commands.reset()
    store.apply({ type: MessageType.BOARD, cells: [{ alive: true }], frame: 2 })
    expect(store.getBoard().cells).toBeNull()
    expect(send).toHaveBeenCalledWith({ resetBoard: true })
  })

  it('lets boards through again once the player starts the clock', () => {
    const { store, commands } = setup()
    commands.setRunning(false)
    commands.setRunning(true)
    store.apply({ type: MessageType.BOARD, cells: [], frame: 5 })
    expect(store.getFrame().frame).toBe(5)
  })
})
