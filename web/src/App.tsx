import { useState } from 'react'
import { Board } from './components/Board.tsx'
import { PatternToolbar } from './components/PatternToolbar.tsx'
import { patterns, type PatternName } from './components/patterns.ts'
import { Presence } from './components/Presence.tsx'
import { useGrid } from './hooks/useGrid.ts'
import { usePresence } from './hooks/usePresence.ts'

function App() {
  const [selected, setSelected] = useState<PatternName>(patterns[0])
  const { cells, place } = useGrid()
  const colours = usePresence()

  return (
    <main className="relative h-svh overflow-hidden bg-paper font-sans text-navy">
      <Presence colours={colours} />
      <PatternToolbar selected={selected} onSelect={setSelected} />
      <Board cells={cells} onPlace={place} />
    </main>
  )
}

export default App
