import { create, toJson, type MessageInitShape } from '@bufbuild/protobuf'
import { describe, expect, it } from 'vitest'
import { MessageType, PaceBound, Playback, ServerMessageSchema } from '../gen/life/v1/socket_pb.js'
import { COLS, ROWS } from './grid.ts'
import { readSocketMessage } from './socketMessage.ts'

function serverMessage(init: MessageInitShape<typeof ServerMessageSchema>) {
  return toJson(ServerMessageSchema, create(ServerMessageSchema, init))
}

describe('readSocketMessage', () => {
  it('reads each message by its type', () => {
    const cells = Array.from({ length: COLS * ROWS }, () => ({ alive: false }))
    expect(
      readSocketMessage(
        serverMessage({ type: MessageType.BOARD, width: COLS, height: ROWS, cells, frame: 4 }),
      ),
    ).toEqual({
      type: MessageType.BOARD,
      cells,
      frame: 4,
    })
    expect(
      readSocketMessage(serverMessage({ type: MessageType.PEOPLE, people: ['#112D4E', '#3F72AF'] })),
    ).toEqual({
      type: MessageType.PEOPLE,
      people: ['#112D4E', '#3F72AF'],
    })
    expect(readSocketMessage(serverMessage({ type: MessageType.YOU, colour: '#112D4E' }))).toEqual({
      type: MessageType.YOU,
      colour: '#112D4E',
    })
    expect(readSocketMessage(serverMessage({ type: MessageType.RESET, colour: '#8C6A21' }))).toEqual({
      type: MessageType.RESET,
      colour: '#8C6A21',
    })
    expect(readSocketMessage(serverMessage({ type: MessageType.ENTERED, colour: '#3F72AF' }))).toEqual({
      type: MessageType.ENTERED,
      colour: '#3F72AF',
    })
    expect(readSocketMessage(serverMessage({ type: MessageType.EXITED, colour: '#3F72AF' }))).toEqual({
      type: MessageType.EXITED,
      colour: '#3F72AF',
    })
    expect(readSocketMessage(serverMessage({ type: MessageType.CHANGED, colour: '#E58700' }))).toEqual({
      type: MessageType.CHANGED,
      colour: '#E58700',
    })
    expect(
      readSocketMessage(
        serverMessage({ type: MessageType.CLOCK, pace: 20, playback: Playback.STOPPED }),
      ),
    ).toEqual({
      type: MessageType.CLOCK,
      running: false,
      pace: 20,
    })
  })

  it('rejects a message with no type or a bad hex', () => {
    expect(readSocketMessage({ width: COLS, height: ROWS, cells: [] })).toBeNull()
    expect(readSocketMessage(serverMessage({ type: MessageType.PEOPLE, people: ['navy'] }))).toBeNull()
    expect(readSocketMessage({ type: 'MESSAGE_TYPE_RESET', colour: true })).toBeNull()
    expect(readSocketMessage({ resetBoard: true })).toBeNull()
    expect(
      readSocketMessage(serverMessage({ type: MessageType.CLOCK, pace: 0, playback: Playback.RUNNING })),
    ).toBeNull()
    expect(
      readSocketMessage(
        serverMessage({ type: MessageType.CLOCK, pace: PaceBound.MAX + 1, playback: Playback.RUNNING }),
      ),
    ).toBeNull()
  })
})
