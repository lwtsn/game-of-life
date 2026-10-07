const hex = /^#[0-9A-Fa-f]{6}$/

export function readColours(value: unknown): string[] | null {
  if (typeof value !== 'object' || value === null) return null

  const colours = (value as { colours?: unknown }).colours
  if (!Array.isArray(colours)) return null
  if (!colours.every((colour) => typeof colour === 'string' && hex.test(colour))) {
    return null
  }

  return colours
}
