// Warm palette matching the store theme. Colour is picked deterministically
// from the category name so the same category always gets the same colour.
const palette = [
  'bg-amber-100 text-amber-900',
  'bg-lime-100 text-lime-900',
  'bg-rose-100 text-rose-900',
  'bg-sky-100 text-sky-900',
  'bg-orange-100 text-orange-900',
  'bg-emerald-100 text-emerald-900',
  'bg-violet-100 text-violet-900',
  'bg-stone-200 text-stone-800',
]

export function categoryColour(category: string) {
  let hash = 0
  for (const ch of category.toLowerCase()) hash = (hash * 31 + ch.charCodeAt(0)) | 0
  return palette[Math.abs(hash) % palette.length]
}
