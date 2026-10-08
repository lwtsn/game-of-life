const hex = /^#[0-9A-Fa-f]{6}$/

export function readReset(value: unknown): string | null {
  if (typeof value !== 'object' || value === null) return null

  const reset = (value as { reset?: unknown }).reset
  if (typeof reset !== 'string' || !hex.test(reset)) return null

  return reset
}
