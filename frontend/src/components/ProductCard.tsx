import type { Product } from '../types'

interface Props {
  product: Product
  onAdd: () => void
  disabled?: boolean
}

export default function ProductCard({ product, onAdd, disabled }: Props) {
  return (
    <div className="bg-white rounded-lg border border-[#e8ddd0] p-4 flex flex-col gap-2">
      <h3 className="font-semibold text-[#3d2b1a]">{product.name}</h3>
      <p className="text-sm text-[#8a6a50] flex-1">{product.description}</p>
      <p className="text-xs text-[#8a6a50]">{product.unit}</p>
      <button
        onClick={onAdd}
        disabled={disabled}
        className="mt-2 bg-[#5c3d1e] text-white rounded px-3 py-1.5 text-sm disabled:opacity-50 hover:bg-[#3d2b1a] transition-colors"
      >
        Add
      </button>
    </div>
  )
}
