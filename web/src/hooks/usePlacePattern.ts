import { useCallback } from 'react'
import { placePattern } from '../api/layout.ts'
import type { Pattern } from '../gen/life/v1/pattern_pb.js'

export function usePlacePattern() {
  return useCallback((pattern: Pattern, x: number, y: number) => {
    void placePattern(pattern, x, y)
  }, [])
}
