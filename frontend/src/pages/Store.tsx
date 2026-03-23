import { useQuery } from '@tanstack/react-query'
import { getProducts } from '../api/products'
import ProductCard from '../components/ProductCard'
import { useBasket } from '../hooks/useBasket'

function SkeletonGrid() {
  return (
    <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-4">
      {Array.from({ length: 6 }).map((_, i) => (
        <div key={i} className="bg-white rounded-lg border border-[#e8ddd0] p-4 h-40 animate-pulse" />
      ))}
    </div>
  )
}

export default function Store() {
  const { data: products, isLoading, isError, refetch } = useQuery({
    queryKey: ['products'],
    queryFn: getProducts,
  })
  const { addItem, isValidating } = useBasket()

  if (isLoading) return <SkeletonGrid />

  if (isError) {
    return (
      <div className="text-center py-16">
        <p className="text-[#8a6a50] mb-4">Could not load products. Try again.</p>
        <button
          onClick={() => refetch()}
          className="bg-[#5c3d1e] text-white rounded px-4 py-2 text-sm"
        >
          Retry
        </button>
      </div>
    )
  }

  const available = products?.filter((p) => p.available) ?? []

  if (available.length === 0) {
    return <p className="text-center py-16 text-[#8a6a50]">No products available yet.</p>
  }

  return (
    <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-4">
      {available.map((p) => (
        <ProductCard
          key={p.id}
          product={p}
          onAdd={() => addItem(p)}
          disabled={isValidating}
        />
      ))}
    </div>
  )
}
