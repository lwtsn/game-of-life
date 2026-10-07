import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import App from './App.tsx'
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
    expect(screen.getByRole('button', { name: 'Glider' })).toBeTruthy()
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

describe('fitCell', () => {
  it('grows the 80 by 50 grid to fill a desktop frame', () => {
    const cell = fitCell(1200, 760)
    expect(cell).toBeGreaterThanOrEqual(10)
  })
})
