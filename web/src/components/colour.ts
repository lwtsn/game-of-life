export type Hsv = { h: number; s: number; v: number }

const hexColour = /^#([0-9A-Fa-f]{6})$/

export function hsvToHex(h: number, s: number, v: number) {
  const channel = v * s
  const x = channel * (1 - Math.abs(((h / 60) % 2) - 1))
  const match = v - channel
  let red = 0
  let green = 0
  let blue = 0
  if (h < 60) {
    red = channel
    green = x
  } else if (h < 120) {
    red = x
    green = channel
  } else if (h < 180) {
    green = channel
    blue = x
  } else if (h < 240) {
    green = x
    blue = channel
  } else if (h < 300) {
    red = x
    blue = channel
  } else {
    red = channel
    blue = x
  }
  const byte = (value: number) => Math.round((value + match) * 255).toString(16).padStart(2, '0')
  return `#${byte(red)}${byte(green)}${byte(blue)}`.toUpperCase()
}

export function hexToHsv(hex: string): Hsv | null {
  const found = hexColour.exec(hex)
  if (!found) return null
  const value = Number.parseInt(found[1], 16)
  const red = ((value >> 16) & 255) / 255
  const green = ((value >> 8) & 255) / 255
  const blue = (value & 255) / 255
  const max = Math.max(red, green, blue)
  const min = Math.min(red, green, blue)
  const delta = max - min
  let hue = 0
  if (delta !== 0) {
    if (max === red) hue = ((green - blue) / delta) % 6
    else if (max === green) hue = (blue - red) / delta + 2
    else hue = (red - green) / delta + 4
    hue *= 60
    if (hue < 0) hue += 360
  }
  return { h: hue, s: max === 0 ? 0 : delta / max, v: max }
}
