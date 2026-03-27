import { useParams, Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { getOrder } from '../api/orders'
import StatusBadge from '../components/StatusBadge'
import type { Order } from '../types'

const TERMINAL = new Set<Order['status']>(['fulfilled', 'cancelled'])

export default function OrderStatus() {
  const { id } = useParams<{ id: string }>()

  const { data: order, isLoading, isError } = useQuery({
    queryKey: ['order', id],
    queryFn: () => getOrder(id!),
    enabled: !!id,
    refetchInterval: (query) => {
      const status = query.state.data?.status
      return status && TERMINAL.has(status) ? false : 10_000
    },
  })

  if (isLoading) {
    return <div className="animate-pulse bg-white rounded-lg h-48 border border-[#e8ddd0]" />
  }

  if (isError) {
    return (
      <div className="text-center py-16">
        <p className="text-[#8a6a50]">Could not load order. Try again.</p>
      </div>
    )
  }

  if (!order) {
    return <p className="text-center py-16 text-[#8a6a50]">Order not found.</p>
  }

  return (
    <div className="max-w-lg mx-auto">
      <div className="flex items-center gap-3 mb-2">
        <h1 className="text-2xl font-bold text-[#3d2b1a]">Order</h1>
        <StatusBadge status={order.status} />
      </div>
      <p className="text-sm text-[#8a6a50] mb-6">
        Placed {new Date(order.created_at).toLocaleDateString()}
      </p>

      <ul className="space-y-3 mb-8">
        {order.items.map((item) => (
          <li
            key={item.product_id}
            className="flex justify-between bg-white rounded-lg border border-[#e8ddd0] p-4"
          >
            <span className="text-[#3d2b1a]">{item.product_name}</span>
            <span className="text-[#8a6a50]">
              {item.unit ? `${item.unit} × ${item.quantity}` : `× ${item.quantity}`}
            </span>
          </li>
        ))}
      </ul>

      <Link to="/orders" className="text-[#5c3d1e] text-sm underline">
        ← Order history
      </Link>
    </div>
  )
}
