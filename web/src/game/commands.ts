import { create, toJson, type JsonValue, type MessageInitShape } from '@bufbuild/protobuf'
import { ClientMessageSchema, Playback } from '../gen/life/v1/socket_pb.js'
import type { GameStore } from './store.ts'

type Action = MessageInitShape<typeof ClientMessageSchema>['action']

// GameCommands is everything a player can ask the server to do.
export type GameCommands = {
  readonly place: (x: number, y: number) => void
  readonly reset: () => void
  readonly chooseColour: (colour: string) => void
  readonly setRunning: (on: boolean) => void
  readonly setPace: (pace: number) => void
}

// createCommands builds typed commands over any JSON sender, so they can be tested without a socket.
export function createCommands(send: (message: JsonValue) => void, store: GameStore): GameCommands {
  const act = (action: Action) => send(toJson(ClientMessageSchema, create(ClientMessageSchema, { action })))

  return {
    place: (x, y) => act({ case: 'point', value: { x, y } }),
    reset: () => {
      store.halt()
      store.clearBoard()
      act({ case: 'resetBoard', value: true })
    },
    chooseColour: (colour) => act({ case: 'colour', value: colour }),
    setRunning: (on) => {
      if (on) store.resume()
      else store.halt()
      act({ case: 'playback', value: on ? Playback.RUNNING : Playback.STOPPED })
    },
    setPace: (pace) => act({ case: 'pace', value: pace }),
  }
}
