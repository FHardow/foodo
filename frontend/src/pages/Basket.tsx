import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { getOrder } from '../api/orders'
import { getProducts } from '../api/products'
import { useBasket } from '../hooks/useBasket'
import { useBasketStore } from '../store/basket'

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
        <p className="text-[#8a6a50] mb-4">Your basket is empty</p>
        <Link to="/" className="text-[#5c3d1e] underline">
          Browse products
        </Link>
      </div>
    )
  }

  if (isLoading) {
    return <div className="animate-pulse bg-white rounded-lg h-48 border border-[#e8ddd0]" />
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
      <h1 className="text-2xl font-bold text-[#3d2b1a] mb-6">Your Basket</h1>

      {order?.items.length === 0 ? (
        <p className="text-[#8a6a50]">
          No items yet.{' '}
          <Link to="/" className="underline text-[#5c3d1e]">
            Add some
          </Link>
        </p>
      ) : (
        <ul className="space-y-4 mb-8">
          {order?.items.map((item) => {
            const unit = item.unit || productMap.get(item.product_id)?.unit
            return (
              <li
                key={item.product_id}
                className="flex items-center gap-4 bg-white rounded-lg border border-[#e8ddd0] p-4"
              >
                <div className="flex-1">
                  <p className="font-medium text-[#3d2b1a]">{item.product_name}</p>
                </div>
                <div className="flex items-center gap-2">
                  <button
                    onClick={() =>
                      item.quantity === 1
                        ? removeItem(item.product_id)
                        : updateQuantity(item.product_id, item.quantity - 1)
                    }
                    className="w-7 h-7 rounded border border-[#e8ddd0] text-[#5c3d1e] hover:border-[#5c3d1e]"
                  >
                    −
                  </button>
                  <span className="text-center text-[#3d2b1a] min-w-[3rem]">
                    {unit ? `${unit} ${item.quantity}` : item.quantity}
                  </span>
                  <button
                    onClick={() => updateQuantity(item.product_id, item.quantity + 1)}
                    className="w-7 h-7 rounded border border-[#e8ddd0] text-[#5c3d1e] hover:border-[#5c3d1e]"
                  >
                    +
                  </button>
                  <button
                    onClick={() => removeItem(item.product_id)}
                    className="ml-2 text-red-400 text-sm hover:text-red-600"
                  >
                    Remove
                  </button>
                </div>
              </li>
            )
          })}
        </ul>
      )}

      {confirmError && <p className="text-red-600 mb-4 text-sm">{confirmError}</p>}

      <button
        onClick={handleConfirm}
        disabled={confirming || !order?.items.length}
        className="w-full bg-[#5c3d1e] text-white rounded-lg py-3 font-medium disabled:opacity-50 hover:bg-[#3d2b1a] transition-colors"
      >
        {confirming ? 'Placing order…' : 'Place Order'}
      </button>

      <div className="mt-8 text-center sm:hidden">
        <Link to="/orders" className="text-[#5c3d1e] text-sm underline">
          View order history
        </Link>
      </div>
    </div>
  )
}
