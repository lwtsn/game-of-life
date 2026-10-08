import { describe, expect, it } from 'vitest'
import { appliedCount } from './applied.ts'

describe('appliedCount', () => {
  it('counts the stamps inside the last second', () => {
    const stamps = [100, 400, 900]
    expect(appliedCount(stamps, 1000)).toBe(3)
    expect(appliedCount(stamps, 1400)).toBe(2)
    expect(stamps).toEqual([400, 900])
  })

  it('returns zero once the window is empty', () => {
    expect(appliedCount([10, 20], 2000)).toBe(0)
  })
})
