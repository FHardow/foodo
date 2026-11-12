import { categoryColour } from '../lib/categoryColour'

interface Props {
  category: string
  className?: string
  /** Shows a × button that calls this when clicked. */
  onRemove?: () => void
  /** Draws a ring around the pill, e.g. when it's about to be removed. */
  highlighted?: boolean
}

export default function CategoryPill({ category, className = '', onRemove, highlighted }: Props) {
  return (
    <span
      className={`inline-flex items-center gap-1 rounded-full px-2.5 py-0.5 text-xs font-medium shadow-sm transition-shadow ${categoryColour(category)} ${
        highlighted ? 'ring-2 ring-[#5c3d1e] ring-offset-1' : ''
      } ${className}`}
    >
      {category}
      {onRemove && (
        <button
          type="button"
          onClick={(e) => {
            e.stopPropagation()
            onRemove()
          }}
          aria-label={`Remove ${category}`}
          className="-mr-1 ml-0.5 rounded-full w-4 h-4 inline-flex items-center justify-center opacity-60 hover:opacity-100 hover:bg-black/10"
        >
          ×
        </button>
      )}
    </span>
  )
}
