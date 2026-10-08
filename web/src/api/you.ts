const hex = /^#[0-9A-Fa-f]{6}$/

export function readYou(value: unknown): string | null {
  if (typeof value !== 'object' || value === null) return null
  const you = (value as { you?: unknown }).you
  if (typeof you !== 'string' || !hex.test(you)) return null
  return you
}
