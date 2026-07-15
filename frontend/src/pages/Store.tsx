import { useQuery } from '@tanstack/react-query'
import { getProducts } from '../api/products'
import ProductCard from '../components/ProductCard'
import { useBasket } from '../hooks/useBasket'
import { Card } from '../components/ui/card'
import { Button } from '../components/ui/button'

function SkeletonGrid() {
  return (
    <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-4">
      {Array.from({ length: 6 }).map((_, i) => (
        <Card key={i} className="h-40 animate-pulse" />
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
        <p className="text-muted-foreground mb-4">Could not load products. Try again.</p>
        <Button onClick={() => refetch()}>Retry</Button>
      </div>
    )
  }

  const available = products?.filter((p) => p.available) ?? []

  if (available.length === 0) {
    return <p className="text-center py-16 text-muted-foreground">No products available yet.</p>
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
