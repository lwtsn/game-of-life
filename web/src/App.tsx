import { useEffect, useRef, useState } from 'react'
import { Toaster } from 'sonner'
import { Board } from './components/Board.tsx'
import { ColourWheel } from './components/ColourWheel.tsx'
import { PatternToolbar } from './components/PatternToolbar.tsx'
import { Presence } from './components/Presence.tsx'
import { useGrid } from './hooks/useGrid.ts'
import { usePlacePattern } from './hooks/usePlacePattern.ts'
import { usePresence } from './hooks/usePresence.ts'
import { useResetNotice } from './hooks/useResetNotice.ts'
import { useYou } from './hooks/useYou.ts'

function App() {
  const { cells, place, reset, chooseColour } = useGrid()
  const placePattern = usePlacePattern()
  const colours = usePresence()
  const you = useYou()
  useResetNotice()
  const [open, setOpen] = useState(false)
  const [draft, setDraft] = useState<string | null>(null)
  const dock = useRef<HTMLDivElement>(null)
  const shown = draft ?? you ?? '#DBE2EF'

  useEffect(() => {
    if (!open) return
    function onPointerDown(event: PointerEvent) {
      if (!dock.current?.contains(event.target as Node)) setOpen(false)
    }
    function onKey(event: KeyboardEvent) {
      if (event.key === 'Escape') setOpen(false)
    }
    document.addEventListener('pointerdown', onPointerDown)
    document.addEventListener('keydown', onKey)
    return () => {
      document.removeEventListener('pointerdown', onPointerDown)
      document.removeEventListener('keydown', onKey)
    }
  }, [open])

  return (
    <main className="relative h-svh overflow-hidden bg-paper font-sans text-navy">
      <Presence colours={colours} />
      <PatternToolbar onPlace={placePattern} />
      <Board cells={cells} onPlace={place} />
      <div ref={dock} className="fixed bottom-4 right-4 z-40 flex flex-col items-end gap-2">
        {open && (
          <ColourWheel
            colour={you ?? '#112D4E'}
            onPreview={setDraft}
            onPick={(hex) => {
              setDraft(hex)
              chooseColour(hex)
            }}
          />
        )}
        <div className="flex items-center gap-2">
          <button
            type="button"
            aria-expanded={open}
            onClick={() => setOpen((value) => !value)}
            className="flex items-center gap-2 border border-blue/30 bg-paper px-3 py-2 text-sm text-navy"
          >
            <span className="size-3.5 rounded-full" style={{ backgroundColor: shown }} />
            Colour
          </button>
          <button
            type="button"
            onClick={reset}
            className="border border-blue/30 bg-paper px-3 py-2 text-sm text-navy"
          >
            Reset
          </button>
        </div>
      </div>
      <Toaster
        position="bottom-right"
        offset={{ bottom: 72, right: 16 }}
        mobileOffset={{ bottom: 72, right: 16 }}
      />
    </main>
  )
}

export default App
