import { describe, expect, it } from 'vitest'
import { readArrival } from './arrival.ts'

describe('readArrival', () => {
  it('reads an entrance and an exit', () => {
    expect(readArrival({ entered: '#8C6A21' })).toEqual({
      text: 'entered the game',
      colour: '#8C6A21',
    })
    expect(readArrival({ exited: '#3F72AF' })).toEqual({
      text: 'exited the game',
      colour: '#3F72AF',
    })
  })

  it('ignores a grid frame, a colour list, and a recolour', () => {
    expect(readArrival({ width: 80, height: 50, cells: [] })).toBeNull()
    expect(readArrival({ colours: ['#8C6A21'] })).toBeNull()
    expect(readArrival({ you: '#8C6A21' })).toBeNull()
    expect(readArrival({ entered: 'navy' })).toBeNull()
    expect(readArrival({ exited: true })).toBeNull()
  })
})
