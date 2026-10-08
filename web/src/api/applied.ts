// appliedCount drops stamps older than one second and returns how many remain.
export function appliedCount(stamps: number[], now: number): number {
  const cutoff = now - 1000
  let drop = 0
  while (drop < stamps.length && stamps[drop] < cutoff) drop++
  if (drop > 0) stamps.splice(0, drop)
  return stamps.length
}
