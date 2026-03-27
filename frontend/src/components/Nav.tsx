import { Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { getOrder } from '../api/orders'
import { useBasketStore } from '../store/basket'
import keycloak from '../auth/keycloak'

export default function Nav() {
  const basketOrderId = useBasketStore((s) => s.basketOrderId)
  const { data: order } = useQuery({
    queryKey: ['order', basketOrderId],
    queryFn: () => getOrder(basketOrderId!),
    enabled: !!basketOrderId,
  })

  const itemCount = order?.items.reduce((sum, item) => sum + item.quantity, 0) ?? 0
  const isOwner = keycloak.hasRealmRole('owner')

  return (
    <nav className="bg-white border-b border-[#e8ddd0] px-4 py-3 flex justify-between items-center sticky top-0 z-10">
      <Link to="/" className="font-bold text-[#5c3d1e] text-lg">
        Bread Order
      </Link>
      <div className="flex items-center gap-4">
        <Link to="/" className="hidden sm:block text-sm text-[#5c3d1e] hover:text-[#3d2b1a]">
          Store
        </Link>
        <Link to="/orders" className="hidden sm:block text-sm text-[#5c3d1e] hover:text-[#3d2b1a]">
          History
        </Link>
        {isOwner && (
          <Link to="/admin/products" className="hidden sm:block text-sm text-[#5c3d1e] hover:text-[#3d2b1a]">
            Manage
          </Link>
        )}
        <Link
          to="/basket"
          aria-label={`Basket${itemCount > 0 ? `, ${itemCount} item${itemCount !== 1 ? 's' : ''}` : ''}`}
          className="bg-[#5c3d1e] text-white rounded-full px-3 py-1 text-sm hover:bg-[#3d2b1a] transition-colors"
        >
          🛒{itemCount > 0 ? ` ${itemCount}` : ''}
        </Link>
      </div>
    </nav>
  )
}
