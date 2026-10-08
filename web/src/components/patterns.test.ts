import { describe, expect, it } from 'vitest'
import { cellAtPoint } from './patterns.ts'

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
