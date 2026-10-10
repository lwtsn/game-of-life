import { useEffect, useState } from 'react'
import { listPatterns } from '../api/layout.ts'
import { patternButtonsFrom, type PatternButton } from '../components/patterns.ts'

export type Patterns =
  | { readonly status: 'loading' }
  | { readonly status: 'ready'; readonly buttons: readonly PatternButton[] }
  | { readonly status: 'failed' }

// usePatterns asks the server for its catalogue once, when the component mounts.
export function usePatterns(): Patterns {
  const [patterns, setPatterns] = useState<Patterns>({ status: 'loading' })

  useEffect(() => {
    // StrictMode mounts, unmounts and mounts again in development, so two requests go out.
    // The cleanup marks the first one stale, and only the answer to the live one is stored.
    let live = true
    listPatterns().then(
      (shapes) => {
        if (live) setPatterns({ status: 'ready', buttons: patternButtonsFrom(shapes) })
      },
      () => {
        if (live) setPatterns({ status: 'failed' })
      },
    )
    return () => {
      live = false
    }
  }, [])

  return patterns
}
