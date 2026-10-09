import { create, toJson } from '@bufbuild/protobuf'
import { useEffect, useRef, useState } from 'react'
import { appliedCount } from '../api/applied.ts'
import type { Cell } from '../api/grid.ts'
import { readSocketMessage } from '../api/socketMessage.ts'
import { announcePresence } from '../components/PresenceNotice.tsx'
import { ClientMessageSchema, MessageType, PaceBound, Playback } from '../gen/life/v1/socket_pb.js'
import { useGameSocket } from './useGameSocket.ts'

export function useLiveGame() {
  const { lastJsonMessage, sendJsonMessage, connection, reconnect } = useGameSocket()
  const [cells, setCells] = useState<Cell[] | null>(null)
  const [people, setPeople] = useState<string[]>([])
  const [you, setYou] = useState<string | null>(null)
  const [running, setRunning] = useState(true)
  const [pace, setPace] = useState(PaceBound.MIN)
  const [frame, setFrame] = useState(0)
  const [applied, setApplied] = useState(0)
  // StrictMode runs this effect twice in development. The same message must not be counted or toasted twice.
  const seen = useRef<unknown>(undefined)
  const stamps = useRef<number[]>([])
  // Set on this page's Stop or Reset. Boards already buffered in the browser are ignored until the server confirms the clock is stopped.
  const halting = useRef(false)

  function halt() {
    halting.current = true
    stamps.current = []
    setApplied(0)
    setRunning(false)
  }

  useEffect(() => {
    if (Object.is(seen.current, lastJsonMessage)) return
    seen.current = lastJsonMessage
    const message = readSocketMessage(lastJsonMessage)
    if (!message) return

    switch (message.type) {
      case MessageType.BOARD:
        if (halting.current) return
        stamps.current.push(performance.now())
        setApplied(appliedCount(stamps.current, performance.now()))
        setFrame(message.frame)
        setCells(message.cells)
        return
      case MessageType.PEOPLE:
        setPeople(message.people)
        return
      case MessageType.YOU:
        setYou(message.colour)
        return
      case MessageType.RESET:
        announcePresence(message.colour, 'reset the board')
        return
      case MessageType.ENTERED:
        announcePresence(message.colour, 'entered the game')
        return
      case MessageType.EXITED:
        announcePresence(message.colour, 'exited the game')
        return
      case MessageType.CHANGED:
        announcePresence(message.colour, 'changed colour')
        return
      case MessageType.CLOCK:
        if (halting.current && message.running) return
        setRunning(message.running)
        setPace(message.pace)
        if (!message.running) {
          stamps.current = []
          setApplied(0)
          halting.current = false
        }
        return
    }
  }, [lastJsonMessage])

  useEffect(() => {
    const id = window.setInterval(() => {
      setApplied(appliedCount(stamps.current, performance.now()))
    }, 250)
    return () => window.clearInterval(id)
  }, [])

  return {
    connection,
    reconnect,
    cells,
    people,
    you,
    place: (x: number, y: number) =>
      sendJsonMessage(
        toJson(
          ClientMessageSchema,
          create(ClientMessageSchema, { action: { case: 'point', value: { x, y } } }),
        ),
      ),
    reset: () => {
      halt()
      setCells(null)
      sendJsonMessage(
        toJson(ClientMessageSchema, create(ClientMessageSchema, { action: { case: 'resetBoard', value: true } })),
      )
    },
    chooseColour: (colour: string) =>
      sendJsonMessage(
        toJson(ClientMessageSchema, create(ClientMessageSchema, { action: { case: 'colour', value: colour } })),
      ),
    running,
    pace,
    frame,
    applied,
    setRunning: (on: boolean) => {
      if (on) halting.current = false
      else halt()
      sendJsonMessage(
        toJson(
          ClientMessageSchema,
          create(ClientMessageSchema, {
            action: { case: 'playback', value: on ? Playback.RUNNING : Playback.STOPPED },
          }),
        ),
      )
    },
    setPace: (next: number) =>
      sendJsonMessage(
        toJson(ClientMessageSchema, create(ClientMessageSchema, { action: { case: 'pace', value: next } })),
      ),
  }
}
