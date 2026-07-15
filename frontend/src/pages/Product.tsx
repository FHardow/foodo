import { useParams, useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { ArrowLeft } from 'lucide-react'
import { getProduct } from '../api/products'
import { useBasket } from '../hooks/useBasket'
import { Card } from '../components/ui/card'
import { Button } from '../components/ui/button'

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
        <Card className="overflow-hidden">
          <div className="w-full h-64 bg-secondary" />
          <div className="p-6 space-y-3">
            <div className="h-6 bg-secondary rounded w-1/2" />
            <div className="h-4 bg-secondary rounded w-full" />
            <div className="h-4 bg-secondary rounded w-2/3" />
          </div>
        </Card>
      </div>
    )
  }

  if (isError || !product) {
    return (
      <div className="text-center py-16">
        <p className="text-muted-foreground mb-4">Product not found.</p>
        <Button variant="link" onClick={() => navigate('/')}>
          Back to store
        </Button>
      </div>
    )
  }

  return (
    <div className="max-w-lg mx-auto mt-4">
      <Button
        variant="ghost"
        size="sm"
        onClick={() => navigate(-1)}
        className="mb-4 text-muted-foreground hover:text-primary"
      >
        <ArrowLeft className="h-4 w-4" />
        Back
      </Button>

      <Card className="overflow-hidden">
        {product.image_url ? (
          <img
            src={`${BASE_URL}${product.image_url}`}
            alt={product.name}
            className="w-full h-64 object-cover"
          />
        ) : (
          <div className="w-full h-64 bg-secondary" />
        )}

        <div className="p-6 space-y-4">
          <div>
            <h1 className="text-2xl font-bold text-foreground">{product.name}</h1>
            <span className="text-xs text-muted-foreground uppercase tracking-wide">{product.unit}</span>
          </div>

          <p className="text-primary leading-relaxed">{product.description}</p>

          <div className="flex items-center gap-2 text-sm">
            <span
              className={`inline-block w-2 h-2 rounded-full ${product.available ? 'bg-green-500' : 'bg-muted'}`}
            />
            <span className="text-muted-foreground">
              {product.available ? 'Available' : 'Not available'}
            </span>
          </div>

          {product.available && (
            <Button onClick={() => addItem(product)} disabled={isValidating} className="w-full">
              Add to basket
            </Button>
          )}
        </div>
      </Card>
    </div>
  )
}
