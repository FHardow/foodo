import { useParams, Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { getOrder } from '../api/orders'
import StatusBadge from '../components/StatusBadge'
import { Card } from '../components/ui/card'
import type { Order } from '../types'

const TERMINAL = new Set<Order['status']>(['finished'])

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
    return <Card className="h-48 animate-pulse" />
  }

  if (isError) {
    return (
      <div className="text-center py-16">
        <p className="text-muted-foreground">Could not load order. Try again.</p>
      </div>
    )
  }

  if (!order) {
    return <p className="text-center py-16 text-muted-foreground">Order not found.</p>
  }

  return (
    <div className="max-w-lg mx-auto">
      <div className="flex items-center gap-3 mb-2">
        <h1 className="text-2xl font-bold text-foreground">Order</h1>
        <StatusBadge status={order.status} />
      </div>
      <p className="text-sm text-muted-foreground mb-6">
        Placed {new Date(order.created_at).toLocaleDateString()}
      </p>

      <ul className="space-y-3 mb-8">
        {order.items.map((item) => (
          <li key={item.product_id}>
            <Card className="flex items-center justify-between p-4">
              <span className="text-foreground">{item.product_name}</span>
              <span className="text-muted-foreground">
                {item.unit ? `${item.unit} × ${item.quantity}` : `× ${item.quantity}`}
              </span>
            </Card>
          </li>
        ))}
      </ul>

      <Link to="/orders" className="text-primary text-sm underline">
        ← Order history
      </Link>
    </div>
  )
}
