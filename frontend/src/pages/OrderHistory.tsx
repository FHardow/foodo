import { Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { ChevronRight } from 'lucide-react'
import { getOrders } from '../api/orders'
import StatusBadge from '../components/StatusBadge'
import { Card } from '../components/ui/card'

export default function OrderHistory() {
  const { data: orders, isLoading, isError } = useQuery({
    queryKey: ['orders'],
    queryFn: getOrders,
  })

  if (isLoading) {
    return <Card className="h-48 animate-pulse" />
  }

  if (isError) {
    return (
      <div className="text-center py-16">
        <p className="text-muted-foreground">Could not load orders. Try again.</p>
      </div>
    )
  }

  const sorted = [...(orders ?? [])].sort(
    (a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
  )

  return (
    <div className="max-w-lg mx-auto">
      <h1 className="text-2xl font-bold text-foreground mb-6">Order History</h1>

      {sorted.length === 0 ? (
        <p className="text-muted-foreground">
          No orders yet.{' '}
          <Link to="/" className="underline text-primary">
            Place your first order
          </Link>
        </p>
      ) : (
        <ul className="space-y-3">
          {sorted.map((order) => (
            <li key={order.id}>
              <Link to={`/orders/${order.id}`}>
                <Card className="flex items-center justify-between p-4 hover:border-primary transition-colors">
                  <div>
                    <StatusBadge status={order.status} />
                    <p className="text-sm text-muted-foreground mt-1">
                      {new Date(order.created_at).toLocaleDateString()}
                    </p>
                  </div>
                  <ChevronRight className="h-4 w-4 text-primary" />
                </Card>
              </Link>
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
