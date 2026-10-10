import { create } from '@bufbuild/protobuf'
import { renderHook, waitFor } from '@testing-library/react'
import { StrictMode } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { listPatterns } from '../api/layout.ts'
import { Pattern, ShapeSchema } from '../gen/life/v1/pattern_pb.js'
import { usePatterns } from './usePatterns.ts'

vi.mock('../api/layout.ts', () => ({ listPatterns: vi.fn() }))

const list = vi.mocked(listPatterns)

afterEach(() => list.mockReset())

const glider = create(ShapeSchema, {
  pattern: Pattern.GLIDER,
  label: 'Glider',
  cells: [{ x: 1, y: 0 }, { x: 2, y: 1 }, { x: 0, y: 2 }, { x: 1, y: 2 }, { x: 2, y: 2 }],
})

describe('usePatterns', () => {
  it('is loading until the server answers, then holds the buttons', async () => {
    list.mockResolvedValue([glider])
    const { result } = renderHook(() => usePatterns())
    expect(result.current).toEqual({ status: 'loading' })

    await waitFor(() => expect(result.current.status).toBe('ready'))
    if (result.current.status !== 'ready') throw new Error('not ready')
    expect(result.current.buttons.map((item) => item.label)).toEqual(['Glider'])
    expect(result.current.buttons[0].width).toBe(3)
  })

  it('reports a failed request', async () => {
    list.mockRejectedValue(new Error('offline'))
    const { result } = renderHook(() => usePatterns())
    await waitFor(() => expect(result.current).toEqual({ status: 'failed' }))
  })

  it('keeps only the live answer when StrictMode mounts twice', async () => {
    let first: (shapes: (typeof glider)[]) => void = () => {}
    list
      .mockImplementationOnce(() => new Promise((resolve) => (first = resolve)))
      .mockResolvedValueOnce([glider])
    const { result } = renderHook(() => usePatterns(), { wrapper: StrictMode })

    await waitFor(() => expect(result.current.status).toBe('ready'))
    first([])
    await Promise.resolve()
    if (result.current.status !== 'ready') throw new Error('not ready')
    expect(result.current.buttons).toHaveLength(1)
  })
})
