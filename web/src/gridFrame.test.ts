import { describe, expect, it } from 'vitest'
import { COLS, ROWS } from './components/patterns.ts'
import { readGridFrame } from './gridFrame.ts'

describe('readGridFrame', () => {
  it('accepts a full 80 by 50 frame', () => {
    const cells = Array.from({ length: COLS * ROWS }, (_, index) =>
      index === 3 ? 1 : 0,
    )
    expect(readGridFrame({ width: COLS, height: ROWS, cells })).toEqual({
      width: COLS,
      height: ROWS,
      cells,
    })
  })

  it('rejects a short frame', () => {
    expect(readGridFrame({ width: COLS, height: ROWS, cells: [1] })).toBeNull()
  })
})
