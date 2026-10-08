import { useState } from 'react'
import { PaceBound } from '../gen/life/v1/socket_pb.js'

export function FrameReadout({ frame, applied }: { frame: number; applied: number }) {
  return (
    <p className="text-right text-sm leading-5">
      <span className="block">Frame {frame}</span>
      <span className="block text-blue">{applied} updates/s</span>
    </p>
  )
}

export function ClockControls({
  running,
  pace,
  onRunning,
  onPace,
}: {
  running: boolean
  pace: number
  onRunning: (running: boolean) => void
  onPace: (pace: number) => void
}) {
  const [dragged, setDragged] = useState<number | null>(null)
  const shown = dragged ?? pace

  return (
    <div className="flex min-w-0 flex-1 items-center gap-3">
      <label className="flex min-w-0 flex-1 items-center gap-3 text-sm">
        Speed
        <input
          type="range"
          min={PaceBound.MIN}
          max={PaceBound.MAX}
          step={1}
          value={shown}
          onChange={(event) => {
            const next = Number(event.target.value)
            setDragged(next)
            onPace(next)
          }}
          onPointerUp={() => setDragged(null)}
          onKeyUp={() => setDragged(null)}
          onBlur={() => setDragged(null)}
          className="min-w-0 flex-1 accent-navy"
        />
      </label>
      <span className="w-8 text-right text-sm tabular-nums">{shown}</span>
      <button type="button" onClick={() => onRunning(!running)} className="border border-navy bg-paper px-4 py-2 text-sm">
        {running ? 'Stop' : 'Start'}
      </button>
    </div>
  )
}
