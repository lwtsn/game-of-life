import type { DragEvent } from 'react'
import type { Pattern } from '../gen/life/v1/pattern_pb.js'
import { patternButtons } from './patterns.ts'

function alive(cells: readonly (readonly number[])[], x: number, y: number) {
  return cells.some(([cellX, cellY]) => cellX === x && cellY === y)
}

function hideDragImage(event: DragEvent<HTMLButtonElement>) {
  const blank = document.createElement('canvas')
  blank.width = 1
  blank.height = 1
  event.dataTransfer.setDragImage(blank, 0, 0)
}

export function PatternToolbar({ onDragPattern }: { onDragPattern: (pattern: Pattern | null) => void }) {
  return (
    <aside
      aria-label="Patterns"
      className="flex min-h-0 gap-2 overflow-x-auto border-b border-mist bg-paper px-4 py-3 md:flex-col md:items-center md:overflow-y-auto md:border-r md:border-b-0"
    >
      <p className="hidden text-[11px] font-semibold tracking-[0.18em] text-blue uppercase md:block">Patterns</p>
      {patternButtons.map((item) => (
        <button
          key={item.pattern}
          type="button"
          draggable
          onDragStart={(event) => {
            event.dataTransfer.setData('text/plain', String(item.pattern))
            event.dataTransfer.effectAllowed = 'copy'
            hideDragImage(event)
            onDragPattern(item.pattern)
          }}
          onDragEnd={() => onDragPattern(null)}
          className="flex cursor-grab flex-col items-center gap-1 border border-mist p-2 text-sm active:cursor-grabbing"
        >
          <span
            aria-hidden="true"
            className="grid gap-px bg-mist"
            style={{ gridTemplateColumns: `repeat(${item.width}, 12px)` }}
          >
            {Array.from({ length: item.width * item.height }, (_, index) => {
              const x = index % item.width
              const y = Math.floor(index / item.width)
              return <span key={index} className={alive(item.cells, x, y) ? 'size-3 bg-navy' : 'size-3 bg-paper'} />
            })}
          </span>
          {item.label}
        </button>
      ))}
    </aside>
  )
}
