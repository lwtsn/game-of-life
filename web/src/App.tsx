import { useEffect, useState } from 'react'
import { Board } from './components/Board.tsx'
import { PatternToolbar } from './components/PatternToolbar.tsx'
import { patterns, type PatternName } from './components/patterns.ts'
import { Presence } from './components/Presence.tsx'
import { readGridFrame } from './gridFrame.ts'
import { useGameSocket } from './useGameSocket.ts'

function App() {
  const [selected, setSelected] = useState<PatternName>(patterns[0])
  const { lastJsonMessage } = useGameSocket()
  const [cells, setCells] = useState<number[] | null>(null)

  useEffect(() => {
    const frame = readGridFrame(lastJsonMessage)
    if (frame) setCells(frame.cells)
  }, [lastJsonMessage])

  return (
    <main className="relative h-svh overflow-hidden bg-paper font-sans text-navy">
      <Presence />
      <PatternToolbar selected={selected} onSelect={setSelected} />
      <Board cells={cells} />
    </main>
  )
}

export default App
