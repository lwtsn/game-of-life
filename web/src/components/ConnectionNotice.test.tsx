import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { reconnectDelay } from '../hooks/useGameSocket.ts'
import { ConnectionNotice } from './ConnectionNotice.tsx'

afterEach(() => {
  cleanup()
})

describe('ConnectionNotice', () => {
  it('stays hidden while the socket is open or not configured', () => {
    const { container, rerender } = render(<ConnectionNotice connection="open" onReconnect={vi.fn()} />)
    expect(container.innerHTML).toBe('')
    rerender(<ConnectionNotice connection="off" onReconnect={vi.fn()} />)
    expect(container.innerHTML).toBe('')
  })

  it('says it is reconnecting without offering the button', () => {
    render(<ConnectionNotice connection="reconnecting" onReconnect={vi.fn()} />)
    expect(screen.getByRole('status').textContent).toContain('Reconnecting')
    expect(screen.queryByRole('button', { name: 'Reconnect' })).toBeNull()
  })

  it('offers a reconnect button once retries run out', () => {
    const onReconnect = vi.fn()
    render(<ConnectionNotice connection="lost" onReconnect={onReconnect} />)
    expect(screen.getByRole('alert').textContent).toContain('Lost connection')
    fireEvent.click(screen.getByRole('button', { name: 'Reconnect' }))
    expect(onReconnect).toHaveBeenCalledOnce()
  })
})

describe('reconnectDelay', () => {
  it('doubles from one second and stops at thirty', () => {
    expect(reconnectDelay(0)).toBeGreaterThanOrEqual(1_000)
    expect(reconnectDelay(0)).toBeLessThan(1_500)
    expect(reconnectDelay(3)).toBeGreaterThanOrEqual(8_000)
    expect(reconnectDelay(3)).toBeLessThan(8_500)
    expect(reconnectDelay(20)).toBeGreaterThanOrEqual(30_000)
    expect(reconnectDelay(20)).toBeLessThan(30_500)
  })
})
