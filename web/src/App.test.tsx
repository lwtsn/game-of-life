import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import App from './App.tsx'
import { Board } from './components/Board.tsx'
import { Presence } from './components/Presence.tsx'
import { fitCell } from './components/patterns.ts'

afterEach(() => {
  cleanup()
})

describe('App', () => {
  it('shows connected colours and the pattern rail', () => {
    render(<App />)
    const connected = screen.getByRole('list', { name: 'Connected' })
    expect(connected.querySelectorAll('li')).toHaveLength(0)
    expect(screen.queryByText('Ada')).toBeNull()
    expect(screen.getByRole('complementary', { name: 'Patterns' })).toBeTruthy()
    fireEvent.click(screen.getByRole('button', { name: 'Patterns' }))
    expect(screen.getByRole('button', { name: 'Block' })).toBeTruthy()
    expect(screen.getByRole('button', { name: 'Blinker' })).toBeTruthy()
    expect(screen.getByRole('button', { name: 'Glider' })).toBeTruthy()
    expect(screen.getByRole('button', { name: 'Beacon' })).toBeTruthy()
    expect(screen.queryByRole('button', { name: 'Toad' })).toBeNull()
    fireEvent.click(screen.getByRole('button', { name: 'Glider' }))
    fireEvent.click(screen.getByRole('button', { name: 'Colour' }))
    fireEvent.click(screen.getByRole('button', { name: 'Reset' }))
  })
})

describe('Presence', () => {
  it('draws one dot per connected colour', () => {
    const colours = Array.from({ length: 100 }, (_, index) => {
      const hex = index.toString(16).padStart(6, '0')
      return `#${hex}`
    })
    render(<Presence colours={colours} />)
    expect(screen.getByRole('list', { name: 'Connected' }).querySelectorAll('li')).toHaveLength(100)
  })
})

describe('Board', () => {
  it('sends the column and row of a clicked cell', () => {
    const onPlace = vi.fn()
    render(<Board cells={null} onPlace={onPlace} />)
    const canvas = screen.getByRole('img', { name: 'Game of Life board' })
    canvas.getBoundingClientRect = () =>
      ({
        x: 0,
        y: 0,
        left: 0,
        top: 0,
        right: 200,
        bottom: 200,
        width: 200,
        height: 200,
        toJSON() {
          return {}
        },
      }) as DOMRect

    fireEvent.pointerDown(canvas, { clientX: 6, clientY: 6, button: 0, pointerId: 1 })
    fireEvent.pointerUp(canvas, { pointerId: 1 })
    fireEvent.pointerDown(canvas, { clientX: 19, clientY: 6, button: 0, pointerId: 1 })
    fireEvent.pointerUp(canvas, { pointerId: 1 })
    fireEvent.pointerDown(canvas, { clientX: 12, clientY: 6, button: 0, pointerId: 1 })
    fireEvent.pointerUp(canvas, { pointerId: 1 })
    fireEvent.pointerMove(canvas, { clientX: 6, clientY: 6, pointerId: 1 })

    expect(onPlace).toHaveBeenCalledTimes(2)
    expect(onPlace).toHaveBeenNthCalledWith(1, 0, 0)
    expect(onPlace).toHaveBeenNthCalledWith(2, 1, 0)
  })

  it('paints each cell a drag crosses', () => {
    const onPlace = vi.fn()
    render(<Board cells={null} onPlace={onPlace} />)
    const canvas = screen.getByRole('img', { name: 'Game of Life board' })
    canvas.getBoundingClientRect = () =>
      ({
        x: 0,
        y: 0,
        left: 0,
        top: 0,
        right: 200,
        bottom: 200,
        width: 200,
        height: 200,
        toJSON() {
          return {}
        },
      }) as DOMRect

    fireEvent.pointerDown(canvas, { clientX: 6, clientY: 6, button: 0, pointerId: 1 })
    fireEvent.pointerMove(canvas, { clientX: 6 + 13 * 4, clientY: 6, pointerId: 1 })
    fireEvent.pointerMove(canvas, { clientX: 6 + 13 * 4, clientY: 6, pointerId: 1 })
    fireEvent.pointerUp(canvas, { pointerId: 1 })

    expect(onPlace.mock.calls).toEqual([
      [0, 0],
      [1, 0],
      [2, 0],
      [3, 0],
      [4, 0],
    ])
  })

  it('names the cells that already have a colour', () => {
    const cells = [
      { alive: true, colour: '#7AA0CE', id: 'player-one' },
      { alive: true, colour: '#3F72AF', id: 'player-two' },
    ]
    render(<Board cells={cells} onPlace={() => {}} />)
    expect(
      screen.getByRole('img', { name: 'Game of Life board, cells 0 1 highlighted' }),
    ).toBeTruthy()
  })
})

describe('fitCell', () => {
  it('grows the 80 by 50 grid to fill a desktop frame', () => {
    const cell = fitCell(1200, 760)
    expect(cell).toBeGreaterThanOrEqual(10)
  })
})
