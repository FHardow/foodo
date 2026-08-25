import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { getOrder } from '../api/orders'
import { getProducts } from '../api/products'
import { useBasket } from '../hooks/useBasket'
import { useBasketStore } from '../store/basket'
import { Card } from '../components/ui/card'
import { Button } from '../components/ui/button'

export default function Basket() {
  const navigate = useNavigate()
  const basketOrderId = useBasketStore((s) => s.basketOrderId)
  const { removeItem, updateQuantity, confirm } = useBasket()
  const [confirmError, setConfirmError] = useState<string | null>(null)
  const [confirming, setConfirming] = useState(false)

  const { data: order, isLoading } = useQuery({
    queryKey: ['order', basketOrderId],
    queryFn: () => getOrder(basketOrderId!),
    enabled: !!basketOrderId,
  })

  const { data: products } = useQuery({
    queryKey: ['products'],
    queryFn: getProducts,
  })

  if (!basketOrderId || (!isLoading && !order)) {
    return (
      <div className="text-center py-16">
        <p className="text-muted-foreground mb-4">Your basket is empty</p>
        <Link to="/" className="text-primary underline">
          Browse products
        </Link>
      </div>
    )
  }

  if (isLoading) {
    return <Card className="h-48 animate-pulse" />
  }

  // productMap used as fallback for items that may lack unit (e.g. older orders)
  const productMap = new Map(products?.map((p) => [p.id, p]) ?? [])

  async function handleConfirm() {
    setConfirmError(null)
    setConfirming(true)
    try {
      const id = await confirm()
      if (id) navigate(`/orders/${id}`)
    } catch {
      setConfirmError('Could not place order. Try again.')
    } finally {
      setConfirming(false)
    }
  }

  return (
    <div className="max-w-lg mx-auto">
      <h1 className="text-2xl font-bold text-foreground mb-6">Your Basket</h1>

      {order?.items.length === 0 ? (
        <p className="text-muted-foreground">
          No items yet.{' '}
          <Link to="/" className="underline text-primary">
            Add some
          </Link>
        </p>
      ) : (
        <ul className="space-y-4 mb-8">
          {order?.items.map((item) => {
            const unit = item.unit || productMap.get(item.product_id)?.unit
            return (
              <li key={item.product_id}>
                <Card className="flex items-center gap-4 p-4">
                  <div className="flex-1">
                    <p className="font-medium text-foreground">{item.product_name}</p>
                  </div>
                  <div className="flex items-center gap-2">
                    <Button
                      variant="outline"
                      size="icon"
                      className="h-7 w-7"
                      onClick={() =>
                        item.quantity === 1
                          ? removeItem(item.product_id)
                          : updateQuantity(item.product_id, item.quantity - 1)
                      }
                    >
                      −
                    </Button>
                    <span className="text-center text-foreground min-w-[3rem]">
                      {item.quantity}
                    </span>
                    <Button
                      variant="outline"
                      size="icon"
                      className="h-7 w-7"
                      onClick={() => updateQuantity(item.product_id, item.quantity + 1)}
                    >
                      +
                    </Button>
                    <Button
                      variant="link"
                      className="ml-2 text-red-400 hover:text-red-600 h-auto p-0"
                      onClick={() => removeItem(item.product_id)}
                    >
                      Remove
                    </Button>
                  </div>
                </Card>
              </li>
            )
          })}
        </ul>
      )}

      {confirmError && <p className="text-red-600 mb-4 text-sm">{confirmError}</p>}

      <Button
        onClick={handleConfirm}
        disabled={confirming || !order?.items.length}
        className="w-full py-3"
        size="lg"
      >
        {confirming ? 'Placing order…' : 'Place Order'}
      </Button>

      <div className="mt-8 text-center sm:hidden">
        <Link to="/orders" className="text-primary text-sm underline">
          View order history
        </Link>
      </div>
    </div>
  )
}
