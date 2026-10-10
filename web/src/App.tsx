import { useEffect, useRef, useState } from 'react'
import { Toaster } from 'sonner'
import { Board } from './components/Board.tsx'
import { ClockControls, FrameReadout } from './components/Clock.tsx'
import { ColourWheel } from './components/ColourWheel.tsx'
import { ConnectionNotice } from './components/ConnectionNotice.tsx'
import { PatternToolbar } from './components/PatternToolbar.tsx'
import { Presence } from './components/Presence.tsx'
import {
  GameProvider,
  useBoard,
  useClock,
  useCommands,
  useConnection,
  useFrame,
  usePeople,
  usePresenceToasts,
} from './game/index.ts'
import type { Pattern } from './gen/life/v1/pattern_pb.js'
import type { PatternButton } from './components/patterns.ts'
import { usePatterns } from './hooks/usePatterns.ts'
import { usePlacePattern } from './hooks/usePlacePattern.ts'

const INK = '#112D4E'

function App() {
  return (
    <GameProvider>
      <Game />
    </GameProvider>
  )
}

// Game lays out the page. It only re-renders when people or the colour picker change;
// the board, clock and frame counter each subscribe to their own slice below.
function Game() {
  usePresenceToasts()
  const { people, you, chooseColour } = usePeople()
  const { reset } = useCommands()
  const patterns = usePatterns()
  const [open, setOpen] = useState(false)
  const [dragPattern, setDragPattern] = useState<Pattern | null>(null)
  const [draft, setDraft] = useState<string | null>(null)
  const picked = useRef<string | null>(null)
  const colour = useRef<HTMLDivElement>(null)
  const shown = draft ?? you ?? '#DBE2EF'

  function closeColour() {
    setDraft(picked.current)
    setOpen(false)
  }

  useEffect(() => {
    if (!open) return
    function onPointerDown(event: PointerEvent) {
      if (!colour.current?.contains(event.target as Node)) closeColour()
    }
    function onKey(event: KeyboardEvent) {
      if (event.key === 'Escape') closeColour()
    }
    document.addEventListener('pointerdown', onPointerDown)
    document.addEventListener('keydown', onKey)
    return () => {
      document.removeEventListener('pointerdown', onPointerDown)
      document.removeEventListener('keydown', onKey)
    }
  }, [open])

  return (
    <main className="grid h-svh grid-rows-[auto_auto_minmax(0,1fr)_auto] overflow-hidden bg-paper font-sans text-navy md:grid-cols-[7.5rem_minmax(0,1fr)] md:grid-rows-[auto_minmax(0,1fr)_auto]">
      <header className="flex items-center justify-between gap-4 border-b border-mist px-4 py-3 md:col-span-2">
        <Presence people={people} />
        <LiveFrameReadout />
      </header>
      <LiveConnectionNotice />
      <PatternToolbar patterns={patterns} onDragPattern={setDragPattern} />
      <div className="h-full min-h-0 min-w-0 md:col-start-2 md:row-start-2">
        <LiveBoard
          patterns={patterns.status === 'ready' ? patterns.buttons : undefined}
          colour={you ?? INK}
          dragPattern={dragPattern}
        />
      </div>
      <footer className="flex flex-wrap items-center gap-3 border-t border-mist px-4 py-3 md:col-span-2">
        <LiveClockControls />
        <div ref={colour} className="relative flex items-center gap-2">
          {open && (
            <div className="absolute right-0 bottom-full z-40 mb-2">
              <ColourWheel
                colour={you ?? INK}
                onPreview={setDraft}
                onPick={(hex) => {
                  picked.current = hex
                  setDraft(hex)
                  chooseColour(hex)
                }}
              />
            </div>
          )}
          <button
            type="button"
            aria-expanded={open}
            onClick={() => {
              if (open) closeColour()
              else setOpen(true)
            }}
            className="flex items-center gap-2 border border-mist bg-paper px-3 py-2 text-sm"
          >
            <span className="size-3.5 rounded-full" style={{ backgroundColor: shown }} />
            Colour
          </button>
          <button type="button" onClick={reset} className="border border-navy bg-navy px-4 py-2 text-sm text-paper">
            Reset
          </button>
        </div>
      </footer>
      <Toaster
        position="bottom-right"
        offset={open ? { bottom: 340, right: 16 } : { bottom: 88, right: 16 }}
        mobileOffset={open ? { bottom: 340, right: 16 } : { bottom: 88, right: 16 }}
      />
    </main>
  )
}

type LiveBoardProps = { patterns?: readonly PatternButton[]; colour: string; dragPattern: Pattern | null }

function LiveBoard({ patterns, colour, dragPattern }: LiveBoardProps) {
  const { cells, place } = useBoard()
  const placePattern = usePlacePattern()
  return (
    <Board
      cells={cells}
      patterns={patterns}
      colour={colour}
      dragPattern={dragPattern}
      onPlace={place}
      onStamp={placePattern}
    />
  )
}

function LiveFrameReadout() {
  const { frame, applied } = useFrame()
  return <FrameReadout frame={frame} applied={applied} />
}

function LiveClockControls() {
  const { running, pace, setRunning, setPace } = useClock()
  return <ClockControls running={running} pace={pace} onRunning={setRunning} onPace={setPace} />
}

function LiveConnectionNotice() {
  const { connection, reconnect } = useConnection()
  return <ConnectionNotice connection={connection} onReconnect={reconnect} />
}

export default App
