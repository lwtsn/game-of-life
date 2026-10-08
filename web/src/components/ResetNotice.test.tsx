import { act, cleanup, fireEvent, render, screen } from '@testing-library/react'
import { toast, Toaster } from 'sonner'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { announceReset } from './ResetNotice.tsx'

afterEach(() => {
  act(() => {
    toast.dismiss()
  })
  cleanup()
  vi.useRealTimers()
})

describe('announceReset', () => {
  it('shows the colour that reset the board', async () => {
    render(
      <>
        <Toaster position="bottom-right" />
        <button type="button" onClick={() => announceReset('#8C6A21')}>
          show
        </button>
      </>,
    )

    fireEvent.click(screen.getByRole('button', { name: 'show' }))

    const notice = await screen.findByText('reset the board')
    const mark = notice.querySelector('span')
    expect(mark).toBeTruthy()
    expect((mark as HTMLElement).style.backgroundColor).toBe('rgb(140, 106, 33)')
  })

  it('leaves after five seconds', async () => {
    vi.useFakeTimers()
    render(
      <>
        <Toaster position="bottom-right" />
        <button type="button" onClick={() => announceReset('#8C6A21')}>
          show
        </button>
      </>,
    )

    fireEvent.click(screen.getByRole('button', { name: 'show' }))
    await act(async () => {
      await vi.advanceTimersByTimeAsync(0)
    })
    expect(screen.getByText('reset the board')).toBeTruthy()

    await act(async () => {
      await vi.advanceTimersByTimeAsync(4999)
    })
    expect(screen.getByText('reset the board')).toBeTruthy()

    await act(async () => {
      await vi.advanceTimersByTimeAsync(300)
    })
    expect(screen.queryByText('reset the board')).toBeNull()
  })
})
