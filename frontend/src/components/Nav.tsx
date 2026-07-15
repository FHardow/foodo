import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { Menu, ShoppingCart } from 'lucide-react'
import { getOrder } from '../api/orders'
import { useBasketStore } from '../store/basket'
import keycloak from '../auth/keycloak'
import { Button } from './ui/button'
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetTrigger, SheetClose } from './ui/sheet'

export default function Nav() {
  const [menuOpen, setMenuOpen] = useState(false)
  const basketOrderId = useBasketStore((s) => s.basketOrderId)
  const { data: order } = useQuery({
    queryKey: ['order', basketOrderId],
    queryFn: () => getOrder(basketOrderId!),
    enabled: !!basketOrderId,
  })

  const itemCount = order?.items.reduce((sum, item) => sum + item.quantity, 0) ?? 0
  const isOwner = keycloak.hasRealmRole('owner')

  const links = [
    { to: '/', label: 'Store' },
    { to: '/orders', label: 'History' },
    ...(isOwner ? [{ to: '/admin/orders', label: 'Orders' }] : []),
    ...(isOwner ? [{ to: '/admin/products', label: 'Products' }] : []),
  ]

  return (
    <nav className="bg-white border-b border-border px-4 py-3 flex justify-between items-center sticky top-0 z-10">
      <div className="flex items-center gap-2">
        <Sheet open={menuOpen} onOpenChange={setMenuOpen}>
          <SheetTrigger asChild>
            <Button variant="ghost" size="icon" className="sm:hidden" aria-label="Open menu">
              <Menu />
            </Button>
          </SheetTrigger>
          <SheetContent side="left">
            <SheetHeader>
              <SheetTitle>Menu</SheetTitle>
            </SheetHeader>
            <div className="flex flex-col gap-1 mt-4">
              {links.map((link) => (
                <SheetClose asChild key={link.to}>
                  <Link to={link.to} className="rounded-md px-3 py-2 text-sm text-foreground hover:bg-accent">
                    {link.label}
                  </Link>
                </SheetClose>
              ))}
            </div>
          </SheetContent>
        </Sheet>
        <Link to="/" className="font-bold text-primary text-lg">
          Foodo
        </Link>
      </div>
      <div className="flex items-center gap-4">
        {links.map((link) => (
          <Link key={link.to} to={link.to} className="hidden sm:block text-sm text-primary hover:text-foreground">
            {link.label}
          </Link>
        ))}
        <Button variant="default" size="sm" className="rounded-full" asChild>
          <Link
            to="/basket"
            aria-label={`Basket${itemCount > 0 ? `, ${itemCount} item${itemCount !== 1 ? 's' : ''}` : ''}`}
          >
            <ShoppingCart className="h-4 w-4" />
            {itemCount > 0 ? itemCount : null}
          </Link>
        </Button>
      </div>
    </nav>
  )
}
