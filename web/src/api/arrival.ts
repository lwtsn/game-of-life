const hex = /^#[0-9A-Fa-f]{6}$/

export function readArrival(value: unknown): { text: string; colour: string } | null {
  if (typeof value !== 'object' || value === null) return null

  const body = value as { entered?: unknown; exited?: unknown }
  if (typeof body.entered === 'string' && hex.test(body.entered)) {
    return { text: 'entered the game', colour: body.entered }
  }
  if (typeof body.exited === 'string' && hex.test(body.exited)) {
    return { text: 'exited the game', colour: body.exited }
  }
  return null
}
