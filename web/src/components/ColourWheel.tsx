import { useEffect, useRef, useState, type PointerEvent } from 'react'
import { hexToHsv, hsvToHex, type Hsv } from './colour.ts'

const SIZE = 168

type ColourWheelProps = {
  colour: string
  onPreview: (hex: string) => void
  onPick: (hex: string) => void
}

export function ColourWheel({ colour, onPreview, onPick }: ColourWheelProps) {
  const parsed = hexToHsv(colour) ?? { h: 210, s: 0.64, v: 0.31 }
  const [hsv, setHsv] = useState<Hsv>(parsed)
  const [source, setSource] = useState(colour)
  const [held, setHeld] = useState<string | null>(null)
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const dragging = useRef(false)
  const latest = useRef(parsed)

  if (held && colour.toUpperCase() === held) setHeld(null)
  if (!held && colour !== source) {
    setSource(colour)
    setHsv(parsed)
  }

  useEffect(() => {
    latest.current = hsv
  }, [hsv])

  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas) return
    const context = canvas.getContext('2d')
    if (!context) return
    const ratio = window.devicePixelRatio || 1
    const pixels = Math.round(SIZE * ratio)
    canvas.width = pixels
    canvas.height = pixels
    const image = context.createImageData(pixels, pixels)
    const radius = pixels / 2
    for (let y = 0; y < pixels; y++) {
      for (let x = 0; x < pixels; x++) {
        const dx = x + 0.5 - radius
        const dy = y + 0.5 - radius
        const saturation = Math.hypot(dx, dy) / radius
        const index = (y * pixels + x) * 4
        if (saturation > 1) {
          image.data[index + 3] = 0
          continue
        }
        let hue = (Math.atan2(dy, dx) * 180) / Math.PI
        if (hue < 0) hue += 360
        const hex = hsvToHex(hue, saturation, hsv.v)
        image.data[index] = Number.parseInt(hex.slice(1, 3), 16)
        image.data[index + 1] = Number.parseInt(hex.slice(3, 5), 16)
        image.data[index + 2] = Number.parseInt(hex.slice(5, 7), 16)
        image.data[index + 3] = 255
      }
    }
    context.putImageData(image, 0, 0)
  }, [hsv.v])

  function pointFrom(event: PointerEvent<HTMLCanvasElement>): Hsv {
    const rect = event.currentTarget.getBoundingClientRect()
    const dx = event.clientX - rect.left - rect.width / 2
    const dy = event.clientY - rect.top - rect.height / 2
    let hue = (Math.atan2(dy, dx) * 180) / Math.PI
    if (hue < 0) hue += 360
    return { h: hue, s: Math.min(1, Math.hypot(dx, dy) / (rect.width / 2)), v: latest.current.v }
  }

  function show(next: Hsv) {
    latest.current = next
    setHsv(next)
    onPreview(hsvToHex(next.h, next.s, next.v))
  }

  function onPointerDown(event: PointerEvent<HTMLCanvasElement>) {
    if (event.button !== 0) return
    dragging.current = true
    if (typeof event.currentTarget.setPointerCapture === 'function') {
      event.currentTarget.setPointerCapture(event.pointerId)
    }
    show(pointFrom(event))
  }

  function onPointerMove(event: PointerEvent<HTMLCanvasElement>) {
    if (!dragging.current) return
    show(pointFrom(event))
  }

  function finish() {
    if (!dragging.current) return
    dragging.current = false
    const next = latest.current
    const hex = hsvToHex(next.h, next.s, next.v)
    setHeld(hex)
    onPick(hex)
  }

  function setBrightness(value: number) {
    show({ ...latest.current, v: value })
  }

  function finishBrightness() {
    const next = latest.current
    const hex = hsvToHex(next.h, next.s, next.v)
    setHeld(hex)
    onPick(hex)
  }

  const radius = (SIZE / 2) * hsv.s
  const radians = (hsv.h * Math.PI) / 180
  const markerX = SIZE / 2 + Math.cos(radians) * radius
  const markerY = SIZE / 2 + Math.sin(radians) * radius

  return (
    <div
      role="dialog"
      aria-label="Choose a colour"
      className="w-[196px] border border-blue/30 bg-paper p-3 text-navy shadow-sm"
    >
      <div className="relative mx-auto" style={{ width: SIZE, height: SIZE }}>
        <canvas
          ref={canvasRef}
          role="img"
          aria-label="Colour wheel"
          className="cursor-crosshair touch-none"
          style={{ width: SIZE, height: SIZE }}
          onPointerDown={onPointerDown}
          onPointerMove={onPointerMove}
          onPointerUp={finish}
          onPointerCancel={finish}
        />
        <span
          className="pointer-events-none absolute size-3.5 -translate-x-1/2 -translate-y-1/2 rounded-full border-2 border-white"
          style={{ left: markerX, top: markerY, boxShadow: '0 0 0 1px #112D4E' }}
        />
      </div>
      <label className="mt-3 block text-sm">
        Brightness
        <input
          type="range"
          min={0}
          max={100}
          aria-label="Brightness"
          className="mt-1 block w-full accent-navy"
          value={Math.round(hsv.v * 100)}
          onChange={(event) => setBrightness(Number(event.target.value) / 100)}
          onPointerUp={finishBrightness}
          onKeyUp={finishBrightness}
        />
      </label>
    </div>
  )
}
