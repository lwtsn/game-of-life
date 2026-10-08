import { act, cleanup, fireEvent, render, screen } from '@testing-library/react'
import { toast, Toaster } from 'sonner'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { announcePresence } from './PresenceNotice.tsx'

afterEach(() => {
  act(() => {
    toast.dismiss()
  })
  cleanup()
  vi.useRealTimers()
})

describe('announcePresence', () => {
  it('shows the colour that entered', async () => {
    render(
      <>
        <Toaster position="bottom-right" />
        <button type="button" onClick={() => announcePresence('#8C6A21', 'entered the game')}>
          show
        </button>
      </>,
    )

    fireEvent.click(screen.getByRole('button', { name: 'show' }))

    const notice = await screen.findByText('entered the game')
    const mark = notice.querySelector('span')
    expect(mark).toBeTruthy()
    expect((mark as HTMLElement).style.backgroundColor).toBe('rgb(140, 106, 33)')
  })

  it('shows the colour that left', async () => {
    render(
      <>
        <Toaster position="bottom-right" />
        <button type="button" onClick={() => announcePresence('#3F72AF', 'exited the game')}>
          show
        </button>
      </>,
    )

    fireEvent.click(screen.getByRole('button', { name: 'show' }))
    expect(await screen.findByText('exited the game')).toBeTruthy()
  })

  it('leaves after five seconds', async () => {
    vi.useFakeTimers()
    render(
      <>
        <Toaster position="bottom-right" />
        <button type="button" onClick={() => announcePresence('#8C6A21', 'entered the game')}>
          show
        </button>
      </>,
    )

    fireEvent.click(screen.getByRole('button', { name: 'show' }))
    await act(async () => {
      await vi.advanceTimersByTimeAsync(0)
    })
    expect(screen.getByText('entered the game')).toBeTruthy()

    await act(async () => {
      await vi.advanceTimersByTimeAsync(4999)
    })
    expect(screen.getByText('entered the game')).toBeTruthy()

    await act(async () => {
      await vi.advanceTimersByTimeAsync(300)
    })
    expect(screen.queryByText('entered the game')).toBeNull()
  })
})
