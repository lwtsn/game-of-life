import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import App from './App.tsx'
import { fitCell } from './components/patterns.ts'

afterEach(() => {
  cleanup()
})

describe('App', () => {
  it('shows connected colours and the pattern rail', () => {
    render(<App />)
    const connected = screen.getByRole('list', { name: 'Connected' })
    expect(connected.querySelectorAll('li')).toHaveLength(3)
    expect(screen.queryByText('Ada')).toBeNull()
    expect(screen.getByRole('complementary', { name: 'Patterns' })).toBeTruthy()
    fireEvent.click(screen.getByRole('button', { name: 'Patterns' }))
    expect(screen.getByRole('button', { name: 'Glider' })).toBeTruthy()
  })
})

describe('fitCell', () => {
  it('grows the 80 by 50 grid to fill a desktop frame', () => {
    const cell = fitCell(1200, 760)
    expect(cell).toBeGreaterThanOrEqual(10)
  })
})
