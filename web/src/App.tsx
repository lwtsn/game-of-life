import { Toaster } from 'sonner'
import { Board } from './components/Board.tsx'
import { PatternToolbar } from './components/PatternToolbar.tsx'
import { Presence } from './components/Presence.tsx'
import { useGrid } from './hooks/useGrid.ts'
import { usePlacePattern } from './hooks/usePlacePattern.ts'
import { usePresence } from './hooks/usePresence.ts'
import { useResetNotice } from './hooks/useResetNotice.ts'

function App() {
  const { cells, place, reset } = useGrid()
  const placePattern = usePlacePattern()
  const colours = usePresence()
  useResetNotice()

  return (
    <main className="relative h-svh overflow-hidden bg-paper font-sans text-navy">
      <Presence colours={colours} />
      <PatternToolbar onPlace={placePattern} />
      <Board cells={cells} onPlace={place} />
      <button
        type="button"
        onClick={reset}
        className="fixed bottom-4 right-4 z-40 border border-blue/30 bg-paper px-3 py-2 text-sm text-navy"
      >
        Reset
      </button>
      <Toaster
        position="bottom-right"
        offset={{ bottom: 72, right: 16 }}
        mobileOffset={{ bottom: 72, right: 16 }}
      />
    </main>
  )
}

export default App
