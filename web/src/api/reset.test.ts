import { describe, expect, it } from 'vitest'
import { readReset } from './reset.ts'

describe('readReset', () => {
  it('reads the colour that cleared the board', () => {
    expect(readReset({ reset: '#8C6A21' })).toBe('#8C6A21')
  })

  it('ignores a grid frame and a click', () => {
    expect(readReset({ width: 80, height: 50, cells: [] })).toBeNull()
    expect(readReset({ reset: true })).toBeNull()
    expect(readReset({ colours: ['#8C6A21'] })).toBeNull()
  })
})
