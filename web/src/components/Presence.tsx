type PresenceProps = {
  people: string[]
}

export function Presence({ people }: PresenceProps) {
  return (
    <ul
      aria-label="Connected"
      className="flex max-h-12 max-w-[50vw] flex-wrap content-start gap-1.5 overflow-y-auto"
    >
      {people.map((colour, index) => (
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
