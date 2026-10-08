import { describe, expect, it } from 'vitest'
import { cellAtPoint, cellNear, cellsAlong, stampOrigin } from './patterns.ts'

describe('cellAtPoint', () => {
  it('returns the cell under the point', () => {
    expect(cellAtPoint(6, 6, 12)).toBe(0)
    expect(cellAtPoint(13 + 6, 6, 12)).toBe(1)
  })

  it('ignores the gap and the space outside the board', () => {
    expect(cellAtPoint(12, 6, 12)).toBeNull()
    expect(cellAtPoint(-1, 6, 12)).toBeNull()
  })
})

describe('stampOrigin', () => {
  it('keeps a shape inside the board', () => {
    expect(stampOrigin(79, 49, 4, 4)).toEqual({ x: 76, y: 46 })
    expect(stampOrigin(3, 2, 2, 2)).toEqual({ x: 3, y: 2 })
  })
})

describe('cellNear', () => {
  it('lands on a cell when the pointer is in the gap', () => {
    expect(cellNear(12, 6, 12)).toBe(0)
  })
})

describe('cellsAlong', () => {
  it('joins two samples so a fast drag does not skip a cell', () => {
    expect(cellsAlong(0, 0, 4, 0)).toEqual([
      [0, 0],
      [1, 0],
      [2, 0],
      [3, 0],
      [4, 0],
    ])
    expect(cellsAlong(1, 1, 1, 1)).toEqual([[1, 1]])
  })
})
