import { describe, expect, it, vi } from 'vitest'
import type { SocketMessage } from '../api/socketMessage.ts'
import { MessageType } from '../gen/life/v1/socket_pb.js'
import { GameStore } from './store.ts'

const board = (frame: number): SocketMessage => ({ type: MessageType.BOARD, cells: [{ alive: true }], frame })
const clock = (running: boolean): SocketMessage => ({ type: MessageType.CLOCK, running, pace: 1 })

function storeAt(start = 0) {
  let now = start
  const store = new GameStore(() => now)
  return { store, advance: (ms: number) => (now += ms) }
}

describe('GameStore', () => {
  it('applies boards and counts how many arrived in the last second', () => {
    const { store, advance } = storeAt()
    store.apply(board(1))
    advance(100)
    store.apply(board(2))
    expect(store.getBoard().cells).toEqual([{ alive: true }])
    expect(store.getFrame()).toEqual({ frame: 2, applied: 2 })

    advance(1_500)
    store.tick()
    expect(store.getFrame().applied).toBe(0)
  })

  it('replaces only the slice that changed', () => {
    const { store } = storeAt()
    store.apply(board(1))
    const before = { board: store.getBoard(), clock: store.getClock(), people: store.getPeople() }

    store.apply({ type: MessageType.PEOPLE, people: ['#112d4e'] })

    expect(store.getPeople()).not.toBe(before.people)
    expect(store.getBoard()).toBe(before.board)
    expect(store.getClock()).toBe(before.clock)
  })

  it('ignores boards in flight after a local stop until the server confirms it', () => {
    const { store } = storeAt()
    store.apply(board(1))
    store.halt()
    expect(store.getClock().running).toBe(false)

    store.apply(board(2))
    store.apply(clock(true))
    expect(store.getFrame().frame).toBe(1)
    expect(store.getClock().running).toBe(false)

    store.apply(clock(false))
    store.apply(board(3))
    expect(store.getFrame().frame).toBe(3)
  })

  it('tells listeners about presence without touching state', () => {
    const { store } = storeAt()
    const listener = vi.fn()
    const notice = vi.fn()
    store.subscribe(listener)
    store.onNotice(notice)

    store.apply({ type: MessageType.ENTERED, colour: '#112d4e' })

    expect(notice).toHaveBeenCalledWith({ colour: '#112d4e', text: 'entered the game' })
    expect(listener).not.toHaveBeenCalled()
  })

  it('stops notifying after unsubscribe', () => {
    const { store } = storeAt()
    const listener = vi.fn()
    const unsubscribe = store.subscribe(listener)
    store.apply(board(1))
    unsubscribe()
    store.apply(board(2))
    expect(listener).toHaveBeenCalledOnce()
  })
})
