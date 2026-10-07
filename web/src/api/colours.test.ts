import { describe, expect, it } from 'vitest'
import { readColours } from './colours.ts'

describe('readColours', () => {
  it('reads a list of hex colours', () => {
    expect(readColours({ colours: ['#112D4E', '#172EA1'] })).toEqual([
      '#112D4E',
      '#172EA1',
    ])
  })

  it('ignores a grid frame', () => {
    expect(readColours({ width: 80, height: 50, cells: [0] })).toBeNull()
  })
})
