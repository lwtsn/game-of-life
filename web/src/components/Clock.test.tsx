import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { PaceBound } from '../gen/life/v1/socket_pb.js'
import { ClockControls, FrameReadout } from './Clock.tsx'

afterEach(() => {
  cleanup()
})

describe('ClockControls', () => {
  it('sets the pace from the slider and stops the clock', () => {
    const onRunning = vi.fn()
    const onPace = vi.fn()
    render(
      <>
        <FrameReadout frame={12} applied={40} />
        <ClockControls running pace={PaceBound.MIN} onRunning={onRunning} onPace={onPace} />
      </>,
    )

    expect(screen.getByText('Frame 12')).toBeTruthy()
    expect(screen.getByText('40 updates/s')).toBeTruthy()
    const slider = screen.getByRole('slider', { name: 'Speed' })
    expect(slider.getAttribute('min')).toBe(String(PaceBound.MIN))
    expect(slider.getAttribute('max')).toBe(String(PaceBound.MAX))
    fireEvent.change(slider, { target: { value: String(PaceBound.MAX) } })
    expect(onPace).toHaveBeenCalledWith(PaceBound.MAX)
    fireEvent.click(screen.getByRole('button', { name: 'Stop' }))
    expect(onRunning).toHaveBeenCalledWith(false)
  })

  it('offers start while the clock is stopped', () => {
    render(<ClockControls running={false} pace={20} onRunning={() => {}} onPace={() => {}} />)
    expect(screen.getByRole('button', { name: 'Start' })).toBeTruthy()
    expect(screen.getByRole('slider', { name: 'Speed' }).getAttribute('value')).toBe('20')
  })
})
