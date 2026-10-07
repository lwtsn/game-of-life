type PresenceProps = {
  colours: string[]
}

export function Presence({ colours }: PresenceProps) {
  return (
    <ul
      aria-label="Connected"
      className="absolute top-4 right-4 z-40 flex max-h-48 max-w-56 flex-wrap content-start justify-end gap-1.5 overflow-y-auto"
    >
      {colours.map((colour, index) => (
        <li
          key={`${colour}-${index}`}
          aria-hidden="true"
          className="size-3.5 rounded-full"
          style={{ backgroundColor: colour }}
        />
      ))}
    </ul>
  )
}
