import { useState } from 'react'
import { patterns, type PatternName } from './patterns.ts'

type PatternToolbarProps = {
  selected: PatternName
  onSelect: (pattern: PatternName) => void
}

export function PatternToolbar({ selected, onSelect }: PatternToolbarProps) {
  const [open, setOpen] = useState(false)

  return (
    <aside
      aria-label="Patterns"
      className={`fixed inset-y-0 left-0 z-30 h-svh overflow-hidden border-r border-blue/30 bg-mist text-navy transition-[width] duration-200 ease-out ${open ? 'w-72' : 'w-11'}`}
      onMouseEnter={() => setOpen(true)}
      onMouseLeave={() => setOpen(false)}
    >
      {open ? (
        <div className="flex h-full flex-col pt-4">
          <p className="px-4 pb-2 text-[11px] font-semibold uppercase tracking-[0.18em]">
            Patterns
          </p>
          <div className="flex min-h-0 flex-1 flex-col gap-1 overflow-y-auto px-2 pb-4">
            {patterns.map((pattern) => {
              const isSelected = pattern === selected
              return (
                <button
                  key={pattern}
                  type="button"
                  aria-pressed={isSelected}
                  onClick={() => onSelect(pattern)}
                  className={`shrink-0 border-l-2 px-3 py-2 text-left text-sm ${isSelected ? 'border-blue bg-blue text-paper' : 'border-transparent hover:bg-paper'}`}
                >
                  {pattern}
                </button>
              )
            })}
          </div>
        </div>
      ) : (
        <button
          type="button"
          aria-expanded={false}
          className="flex h-full w-full items-center justify-center"
          onClick={() => setOpen(true)}
        >
          <span className="text-[11px] font-semibold uppercase tracking-[0.18em] [writing-mode:vertical-rl]">
            Patterns
          </span>
        </button>
      )}
    </aside>
  )
}
