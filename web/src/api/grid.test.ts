import { describe, expect, it } from 'vitest'
import { COLS, readGridFrame, ROWS } from './grid.ts'

describe('readGridFrame', () => {
  it('accepts a full 80 by 50 frame of cells', () => {
    const cells = Array.from({ length: COLS * ROWS }, (_, index) =>
      index === 3
        ? { alive: true, ip: '127.0.0.1', colour: '#3F72AF' }
        : { alive: false },
    )
    expect(readGridFrame({ width: COLS, height: ROWS, cells })).toEqual({
      width: COLS,
      height: ROWS,
      cells,
    })
  })

  it('rejects a short frame', () => {
    expect(readGridFrame({ width: COLS, height: ROWS, cells: [{ alive: true }] })).toBeNull()
  })

  it('rejects integer cells', () => {
    const cells = Array.from({ length: COLS * ROWS }, () => 0)
    expect(readGridFrame({ width: COLS, height: ROWS, cells })).toBeNull()
  })
})
