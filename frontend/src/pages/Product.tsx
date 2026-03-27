import { useParams, useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { getProduct } from '../api/products'
import { useBasket } from '../hooks/useBasket'

const BASE_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'

export default function Product() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { addItem, isValidating } = useBasket()

  const { data: product, isLoading, isError } = useQuery({
    queryKey: ['product', id],
    queryFn: () => getProduct(id!),
    enabled: !!id,
  })

  if (isLoading) {
    return (
      <div className="max-w-lg mx-auto mt-8 animate-pulse">
        <div className="bg-white rounded-lg border border-[#e8ddd0] overflow-hidden">
          <div className="w-full h-64 bg-[#f0e8de]" />
          <div className="p-6 space-y-3">
            <div className="h-6 bg-[#f0e8de] rounded w-1/2" />
            <div className="h-4 bg-[#f0e8de] rounded w-full" />
            <div className="h-4 bg-[#f0e8de] rounded w-2/3" />
          </div>
        </div>
      </div>
    )
  }

  if (isError || !product) {
    return (
      <div className="text-center py-16">
        <p className="text-[#8a6a50] mb-4">Product not found.</p>
        <button
          onClick={() => navigate('/')}
          className="text-sm text-[#5c3d1e] hover:underline"
        >
          Back to store
        </button>
      </div>
    )
  }

  return (
    <div className="max-w-lg mx-auto mt-4">
      <button
        onClick={() => navigate(-1)}
        className="text-sm text-[#8a6a50] hover:text-[#5c3d1e] mb-4 inline-flex items-center gap-1"
      >
        ← Back
      </button>

      <div className="bg-white rounded-lg border border-[#e8ddd0] overflow-hidden">
        {product.image_url ? (
          <img
            src={`${BASE_URL}${product.image_url}`}
            alt={product.name}
            className="w-full h-64 object-cover"
          />
        ) : (
          <div className="w-full h-64 bg-[#f0e8de]" />
        )}

        <div className="p-6 space-y-4">
          <div>
            <h1 className="text-2xl font-bold text-[#3d2b1a]">{product.name}</h1>
            <span className="text-xs text-[#8a6a50] uppercase tracking-wide">{product.unit}</span>
          </div>

          <p className="text-[#5c3d1e] leading-relaxed">{product.description}</p>

          <div className="flex items-center gap-2 text-sm">
            <span
              className={`inline-block w-2 h-2 rounded-full ${product.available ? 'bg-green-500' : 'bg-[#e8ddd0]'}`}
            />
            <span className="text-[#8a6a50]">
              {product.available ? 'Available' : 'Not available'}
            </span>
          </div>

          {product.available && (
            <button
              onClick={() => addItem(product)}
              disabled={isValidating}
              className="w-full bg-[#5c3d1e] text-white rounded px-4 py-2.5 text-sm font-medium hover:bg-[#3d2b1a] disabled:opacity-50 transition-colors"
            >
              Add to basket
            </button>
          )}
        </div>
      </div>
    </div>
  )
}
