import { Toaster } from 'sonner'
import { Board } from './components/Board.tsx'
import { PatternToolbar } from './components/PatternToolbar.tsx'
import { Presence } from './components/Presence.tsx'
import { useGrid } from './hooks/useGrid.ts'
import { usePlacePattern } from './hooks/usePlacePattern.ts'
import { usePresence } from './hooks/usePresence.ts'
import { useResetNotice } from './hooks/useResetNotice.ts'
import { useYou } from './hooks/useYou.ts'

function App() {
  const { cells, place, reset, recolour } = useGrid()
  const placePattern = usePlacePattern()
  const colours = usePresence()
  const you = useYou()
  useResetNotice()

  return (
    <main className="relative h-svh overflow-hidden bg-paper font-sans text-navy">
      <Presence colours={colours} />
      <PatternToolbar onPlace={placePattern} />
      <Board cells={cells} onPlace={place} />
      <div className="fixed bottom-4 right-4 z-40 flex items-center gap-2">
        <button
          type="button"
          onClick={recolour}
          className="flex items-center gap-2 border border-blue/30 bg-paper px-3 py-2 text-sm text-navy"
        >
          <span
            className="size-3.5 rounded-full"
            style={{ backgroundColor: you ?? '#DBE2EF' }}
          />
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
      <Toaster
        position="bottom-right"
        offset={{ bottom: 72, right: 16 }}
        mobileOffset={{ bottom: 72, right: 16 }}
      />
    </main>
  )
}

export default App
