import type { Connection } from '../hooks/useGameSocket.ts'

type ConnectionNoticeProps = {
  connection: Connection
  onReconnect: () => void
}

// ConnectionNotice floats over the board, so it never takes a cell in the page grid.
export function ConnectionNotice({ connection, onReconnect }: ConnectionNoticeProps) {
  if (connection === 'open' || connection === 'off') return null

  const lost = connection === 'lost'
  const message = lost ? 'Lost connection to the game.' : connection === 'reconnecting' ? 'Reconnecting…' : 'Connecting…'

  return (
    <div
      role={lost ? 'alert' : 'status'}
      className="fixed top-3 left-1/2 z-50 flex -translate-x-1/2 items-center gap-3 border border-navy bg-paper px-4 py-2 text-sm text-navy shadow"
    >
      <span>{message}</span>
      {lost && (
        <button type="button" onClick={onReconnect} className="border border-navy bg-navy px-3 py-1 text-paper">
          Reconnect
        </button>
      )}
    </div>
  )
}
