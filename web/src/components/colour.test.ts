import { describe, expect, it } from 'vitest'
import { hexToHsv, hsvToHex } from './colour.ts'

describe('hsvToHex', () => {
  it('turns the corners of the wheel into hex', () => {
    expect(hsvToHex(0, 1, 1)).toBe('#FF0000')
    expect(hsvToHex(120, 1, 1)).toBe('#00FF00')
    expect(hsvToHex(240, 1, 1)).toBe('#0000FF')
    expect(hsvToHex(0, 0, 1)).toBe('#FFFFFF')
    expect(hsvToHex(0, 0, 0)).toBe('#000000')
  })

  it('reads a hex back into hue, saturation, and brightness', () => {
    expect(hexToHsv('#FF0000')).toEqual({ h: 0, s: 1, v: 1 })
    expect(hexToHsv('#00ff00')).toEqual({ h: 120, s: 1, v: 1 })
    expect(hexToHsv('red')).toBeNull()
    expect(hexToHsv('#112D4E')?.v).toBeCloseTo(0x4e / 255, 5)
  })
})
