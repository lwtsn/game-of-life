const colors = ['#112D4E', '#3F72AF', '#7AA0CE']

export function Presence() {
  return (
    <ul
      aria-label="Connected"
      className="absolute top-4 right-4 z-40 flex items-center gap-2"
    >
      {colors.map((color) => (
        <li
          key={color}
          aria-hidden="true"
          className="size-3.5 rounded-full"
          style={{ backgroundColor: color }}
        />
      ))}
    </ul>
  )
}
