import { Board } from './components/Board.tsx'
import { PatternToolbar } from './components/PatternToolbar.tsx'
import { Presence } from './components/Presence.tsx'
import { useGrid } from './hooks/useGrid.ts'
import { usePlacePattern } from './hooks/usePlacePattern.ts'
import { usePresence } from './hooks/usePresence.ts'

function App() {
  const { cells, place } = useGrid()
  const placePattern = usePlacePattern()
  const colours = usePresence()

  return (
    <main className="relative h-svh overflow-hidden bg-paper font-sans text-navy">
      <Presence colours={colours} />
      <PatternToolbar onPlace={placePattern} />
      <Board cells={cells} onPlace={place} />
    </main>
  )
}

export default App
