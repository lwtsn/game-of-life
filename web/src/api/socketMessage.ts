import { fromJson, type JsonValue } from '@bufbuild/protobuf'
import { MessageType, PaceBound, Playback, ServerMessageSchema } from '../gen/life/v1/socket_pb.js'
import { isHex } from './hex.ts'
import { readGridFrame, type Cell } from './grid.ts'

export type SocketMessage =
  | { type: MessageType.BOARD; cells: Cell[]; frame: number }
  | { type: MessageType.PEOPLE; people: string[] }
  | { type: MessageType.YOU; colour: string }
  | { type: MessageType.RESET; colour: string }
  | { type: MessageType.ENTERED; colour: string }
  | { type: MessageType.EXITED; colour: string }
  | { type: MessageType.CHANGED; colour: string }
  | { type: MessageType.CLOCK; running: boolean; pace: number }

export function readSocketMessage(value: unknown): SocketMessage | null {
  if (typeof value !== 'object' || value === null || Array.isArray(value)) return null

  let message
  try {
    message = fromJson(ServerMessageSchema, value as JsonValue)
  } catch {
    return null
  }

  switch (message.type) {
    case MessageType.BOARD: {
      const frame = readGridFrame(value)
      if (!frame || !Number.isInteger(message.frame) || message.frame < 0) return null
      return { type: MessageType.BOARD, cells: frame.cells, frame: message.frame }
    }
    case MessageType.PEOPLE: {
      if (!message.people.every(isHex)) return null
      return { type: MessageType.PEOPLE, people: message.people }
    }
    case MessageType.YOU:
    case MessageType.RESET:
    case MessageType.ENTERED:
    case MessageType.EXITED:
    case MessageType.CHANGED: {
      if (!isHex(message.colour)) return null
      return { type: message.type, colour: message.colour }
    }
    case MessageType.CLOCK: {
      if (!knownPace(message.pace) || !knownPlayback(message.playback)) return null
      return {
        type: MessageType.CLOCK,
        running: message.playback === Playback.RUNNING,
        pace: message.pace,
      }
    }
    default:
      return null
  }
}

function knownPace(pace: number): boolean {
  return Number.isInteger(pace) && pace >= PaceBound.MIN && pace <= PaceBound.MAX
}

function knownPlayback(playback: Playback): boolean {
  return playback === Playback.RUNNING || playback === Playback.STOPPED
}
