import { useNavigate } from 'react-router-dom'
import type { Product } from '../types'
import CategoryPill from './CategoryPill'

const BASE_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'

const MAX_VISIBLE = 3

interface Props {
  product: Product
  onAdd: () => void
  disabled?: boolean
}

export default function ProductCard({ product, onAdd, disabled }: Props) {
  const navigate = useNavigate()
  const categories = product.categories ?? []

  return (
    <div
      className="bg-white rounded-lg border border-[#e8ddd0] overflow-hidden flex flex-col cursor-pointer hover:border-[#5c3d1e] transition-colors"
      onClick={() => navigate(`/products/${product.id}`)}
    >
      <div className="relative">
        {product.image_url ? (
          <img
            src={`${BASE_URL}${product.image_url}`}
            alt={product.name}
            className="w-full h-36 object-cover"
          />
        ) : (
          <div className="w-full h-36 bg-[#f0e8de]" />
        )}
        {categories.length > 0 && (
          <div className="absolute top-2 left-2 right-2 flex flex-wrap gap-1">
            {categories.slice(0, MAX_VISIBLE).map((c) => (
              <CategoryPill key={c} category={c} />
            ))}
            {categories.length > MAX_VISIBLE && (
              <span
                className="inline-block rounded-full px-2 py-0.5 text-xs font-medium bg-white/90 text-[#5c3d1e] shadow-sm"
                title={categories.slice(MAX_VISIBLE).join(', ')}
              >
                +{categories.length - MAX_VISIBLE}
              </span>
            )}
          </div>
        )}
      </div>
      <div className="p-4 flex flex-col gap-2 flex-1">
        <h3 className="font-semibold text-[#3d2b1a]">{product.name}</h3>
        <p className="text-sm text-[#8a6a50] flex-1">{product.description}</p>
        <p className="text-xs text-[#8a6a50]">{product.unit}</p>
        <button
          onClick={(e) => { e.stopPropagation(); onAdd() }}
          disabled={disabled}
          className="mt-2 bg-[#5c3d1e] text-white rounded px-3 py-1.5 text-sm disabled:opacity-50 hover:bg-[#3d2b1a] transition-colors"
        >
          Add
        </button>
      </div>
    </div>
  )
}
