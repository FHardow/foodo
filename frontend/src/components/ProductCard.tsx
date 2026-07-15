import { useNavigate } from 'react-router-dom'
import type { Product } from '../types'
import { Card, CardContent } from './ui/card'
import { Button } from './ui/button'

const BASE_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'

interface Props {
  product: Product
  onAdd: () => void
  disabled?: boolean
}

export default function ProductCard({ product, onAdd, disabled }: Props) {
  const navigate = useNavigate()

  return (
    <Card
      className="overflow-hidden flex flex-col cursor-pointer hover:border-primary transition-colors"
      onClick={() => navigate(`/products/${product.id}`)}
    >
      {product.image_url ? (
        <img
          src={`${BASE_URL}${product.image_url}`}
          alt={product.name}
          className="w-full h-36 object-cover"
        />
      ) : (
        <div className="w-full h-36 bg-secondary" />
      )}
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
