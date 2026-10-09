import { appliedCount } from '../api/applied.ts'
import type { Cell } from '../api/grid.ts'
import type { SocketMessage } from '../api/socketMessage.ts'
import { MessageType, PaceBound } from '../gen/life/v1/socket_pb.js'

export type BoardState = { readonly cells: Cell[] | null }
export type FrameState = { readonly frame: number; readonly applied: number }
export type ClockState = { readonly running: boolean; readonly pace: number }
export type PeopleState = { readonly people: string[]; readonly you: string | null }
export type Notice = { readonly colour: string; readonly text: string }

const noticeText: Partial<Record<MessageType, string>> = {
  [MessageType.RESET]: 'reset the board',
  [MessageType.ENTERED]: 'entered the game',
  [MessageType.EXITED]: 'exited the game',
  [MessageType.CHANGED]: 'changed colour',
}

// GameStore is the client's view of the shared game. It knows nothing about React or sockets:
// messages go in through apply, and each slice is replaced only when it changes,
// so a component that reads one slice never re-renders for another.
export class GameStore {
  #board: BoardState = { cells: null }
  #frame: FrameState = { frame: 0, applied: 0 }
  #clock: ClockState = { running: true, pace: PaceBound.MIN }
  #people: PeopleState = { people: [], you: null }

  #stamps: number[] = []
  // Set by a local Stop or Reset. Boards already in flight are ignored until the server confirms the stop.
  #halting = false

  readonly #listeners = new Set<() => void>()
  readonly #noticeListeners = new Set<(notice: Notice) => void>()
  readonly #now: () => number

  constructor(now: () => number = () => performance.now()) {
    this.#now = now
  }

  readonly getBoard = () => this.#board
  readonly getFrame = () => this.#frame
  readonly getClock = () => this.#clock
  readonly getPeople = () => this.#people

  readonly subscribe = (listener: () => void) => {
    this.#listeners.add(listener)
    return () => {
      this.#listeners.delete(listener)
    }
  }

  readonly onNotice = (listener: (notice: Notice) => void) => {
    this.#noticeListeners.add(listener)
    return () => {
      this.#noticeListeners.delete(listener)
    }
  }

  apply(message: SocketMessage) {
    switch (message.type) {
      case MessageType.BOARD:
        if (this.#halting) return
        this.#stamps.push(this.#now())
        this.#board = { cells: message.cells }
        this.#frame = { frame: message.frame, applied: appliedCount(this.#stamps, this.#now()) }
        break
      case MessageType.PEOPLE:
        this.#people = { ...this.#people, people: message.people }
        break
      case MessageType.YOU:
        this.#people = { ...this.#people, you: message.colour }
        break
      case MessageType.CLOCK:
        if (this.#halting && message.running) return
        this.#clock = { running: message.running, pace: message.pace }
        if (!message.running) {
          this.#halting = false
          this.#resetApplied()
        }
        break
      default: {
        const text = noticeText[message.type]
        if (text) this.#noticeListeners.forEach((listener) => listener({ colour: message.colour, text }))
        return
      }
    }
    this.#emit()
  }

  // halt shows the board as stopped straight away and ignores boards already on their way.
  halt() {
    this.#halting = true
    this.#clock = { ...this.#clock, running: false }
    this.#resetApplied()
    this.#emit()
  }

  resume() {
    this.#halting = false
  }

  clearBoard() {
    this.#board = { cells: null }
    this.#emit()
  }

  // tick lets the applied-per-second count fall when boards stop arriving.
  tick() {
    const applied = appliedCount(this.#stamps, this.#now())
    if (applied === this.#frame.applied) return
    this.#frame = { ...this.#frame, applied }
    this.#emit()
  }

  #resetApplied() {
    this.#stamps = []
    if (this.#frame.applied !== 0) this.#frame = { ...this.#frame, applied: 0 }
  }

  #emit() {
    this.#listeners.forEach((listener) => listener())
  }
}
