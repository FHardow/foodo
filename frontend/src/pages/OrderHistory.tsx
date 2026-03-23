import { Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { getOrders } from '../api/orders'
import StatusBadge from '../components/StatusBadge'

export default function OrderHistory() {
  const { data: orders, isLoading, isError } = useQuery({
    queryKey: ['orders'],
    queryFn: getOrders,
  })

  if (isLoading) {
    return <div className="animate-pulse bg-white rounded-lg h-48 border border-[#e8ddd0]" />
  }

  if (isError) {
    return (
      <div className="text-center py-16">
        <p className="text-[#8a6a50]">Could not load orders. Try again.</p>
      </div>
    )
  }

  const sorted = [...(orders ?? [])].sort(
    (a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
  )

  return (
    <div className="max-w-lg mx-auto">
      <h1 className="text-2xl font-bold text-[#3d2b1a] mb-6">Order History</h1>

      {sorted.length === 0 ? (
        <p className="text-[#8a6a50]">
          No orders yet.{' '}
          <Link to="/" className="underline text-[#5c3d1e]">
            Place your first order
          </Link>
        </p>
      ) : (
        <ul className="space-y-3">
          {sorted.map((order) => (
            <li key={order.id}>
              <Link
                to={`/orders/${order.id}`}
                className="flex items-center justify-between bg-white rounded-lg border border-[#e8ddd0] p-4 hover:border-[#5c3d1e] transition-colors"
              >
                <div>
                  <StatusBadge status={order.status} />
                  <p className="text-sm text-[#8a6a50] mt-1">
                    {new Date(order.created_at).toLocaleDateString()}
                  </p>
                </div>
                <span className="text-[#5c3d1e]">→</span>
              </Link>
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
