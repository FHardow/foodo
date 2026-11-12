import { useNavigate } from 'react-router-dom'
import type { Product } from '../types'
import CategoryPill from './CategoryPill'
import { Card, CardContent } from './ui/card'
import { Button } from './ui/button'

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
    <Card
      className="overflow-hidden flex flex-col cursor-pointer hover:border-primary transition-colors"
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
          <div className="w-full h-36 bg-secondary" />
        )}
        {categories.length > 0 && (
          <div className="absolute top-2 left-2 right-2 flex flex-wrap gap-1">
            {categories.slice(0, MAX_VISIBLE).map((c) => (
              <CategoryPill key={c} category={c} />
            ))}
            {categories.length > MAX_VISIBLE && (
              <span
                className="inline-block rounded-full px-2 py-0.5 text-xs font-medium bg-background/90 text-foreground shadow-sm"
                title={categories.slice(MAX_VISIBLE).join(', ')}
              >
                +{categories.length - MAX_VISIBLE}
              </span>
            )}
          </div>
        )}
      </div>
      <CardContent className="p-4 flex flex-col gap-2 flex-1">
        <h3 className="font-semibold text-foreground">{product.name}</h3>
        <p className="text-sm text-muted-foreground flex-1">{product.description}</p>
        <p className="text-xs text-muted-foreground">{product.unit}</p>
        <Button
          onClick={(e) => { e.stopPropagation(); onAdd() }}
          disabled={disabled}
          size="sm"
          className="mt-2"
        >
          Add
        </Button>
      </CardContent>
    </Card>
  )
}
