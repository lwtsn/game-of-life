import { toast } from 'sonner'

const RESET_TOAST_MS = 5000

export function announceReset(colour: string) {
  const id = toast.custom(
    () => (
      <p className="flex items-center gap-2 border border-blue/30 bg-paper px-3 py-2 text-sm text-navy shadow-sm">
        <span className="size-3.5 shrink-0 rounded-full" style={{ backgroundColor: colour }} />
        reset the board
      </p>
    ),
    { unstyled: true, duration: RESET_TOAST_MS },
  )
  // Sonner pauses its own timer while the pointer is over the toast. The
  // toast slides through the Reset button, so that pause can stick.
  window.setTimeout(() => toast.dismiss(id), RESET_TOAST_MS)
}
