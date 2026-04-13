import type { Product, Order } from '../../src/types'

export const PRODUCTS: Product[] = [
  {
    id: 'prod-1',
    name: 'Sourdough Loaf',
    description: 'Classic tangy sourdough with a crispy crust',
    unit: '1 loaf',
    available: true,
  },
  {
    id: 'prod-2',
    name: 'Croissant',
    description: 'Buttery, flaky pastry',
    unit: '1 piece',
    available: true,
  },
  {
    id: 'prod-3',
    name: 'Baguette',
    description: 'Traditional French baguette',
    unit: '1 stick',
    available: true,
  },
  {
    id: 'prod-4',
    name: 'Rye Bread',
    description: 'Dense and hearty rye loaf',
    unit: '1 loaf',
    available: false,
  },
]

export const BASKET_ORDER: Order = {
  id: 'basket-order-id',
  user_id: 'e2e-mock-user',
  status: 'pending',
  items: [],
  created_at: '2026-04-06T10:00:00Z',
}

export const BASKET_ORDER_WITH_ITEMS: Order = {
  id: 'basket-order-id',
  user_id: 'e2e-mock-user',
  status: 'pending',
  items: [
    { product_id: 'prod-1', product_name: 'Sourdough Loaf', unit: '1 loaf', quantity: 2 },
    { product_id: 'prod-2', product_name: 'Croissant', unit: '1 piece', quantity: 3 },
  ],
  created_at: '2026-04-06T10:00:00Z',
}

export const PLACED_ORDER: Order = {
  id: 'placed-order-id',
  user_id: 'e2e-mock-user',
  status: 'created',
  items: [{ product_id: 'prod-1', product_name: 'Sourdough Loaf', unit: '1 loaf', quantity: 2 }],
  created_at: '2026-04-06T10:00:00Z',
}

export const ALL_ORDERS: Order[] = [
  {
    id: 'order-new-1',
    user_id: 'customer-1',
    user_name: 'Alice',
    status: 'created',
    items: [
      { product_id: 'prod-1', product_name: 'Sourdough Loaf', unit: '1 loaf', quantity: 2 },
    ],
    created_at: '2026-04-06T10:00:00Z',
  },
  {
    id: 'order-new-2',
    user_id: 'customer-2',
    user_name: 'Bob',
    status: 'created',
    items: [
      { product_id: 'prod-3', product_name: 'Baguette', unit: '1 stick', quantity: 1 },
    ],
    created_at: '2026-04-06T09:30:00Z',
  },
  {
    id: 'order-accepted-1',
    user_id: 'customer-3',
    user_name: 'Carol',
    status: 'accepted',
    items: [
      { product_id: 'prod-2', product_name: 'Croissant', unit: '1 piece', quantity: 4 },
    ],
    created_at: '2026-04-06T09:00:00Z',
  },
  {
    id: 'order-ongoing-1',
    user_id: 'customer-4',
    user_name: 'Dave',
    status: 'ongoing',
    items: [
      { product_id: 'prod-1', product_name: 'Sourdough Loaf', unit: '1 loaf', quantity: 1 },
      { product_id: 'prod-2', product_name: 'Croissant', unit: '1 piece', quantity: 2 },
    ],
    created_at: '2026-04-06T08:00:00Z',
  },
  {
    id: 'order-finished-1',
    user_id: 'customer-5',
    user_name: 'Eve',
    status: 'finished',
    items: [
      { product_id: 'prod-3', product_name: 'Baguette', unit: '1 stick', quantity: 2 },
    ],
    created_at: '2026-04-06T07:00:00Z',
  },
]
