const storageKey = 'gol.session'
const sessionPattern = /^[A-Za-z0-9-]{8,64}$/

export function sessionId(): string {
  const existing = window.localStorage.getItem(storageKey)
  if (existing && sessionPattern.test(existing)) return existing

  const id = crypto.randomUUID()
  window.localStorage.setItem(storageKey, id)
  return id
}
