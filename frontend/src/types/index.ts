export type UUID = string

export interface Product {
  id: UUID
  name: string
  description: string
  unit: string
  available: boolean
  image_url?: string
}

export interface OrderItem {
  product_id: UUID
  product_name: string
  unit: string
  quantity: number
  // API also returns unit_price_cents, total_cents — intentionally omitted
}

export interface Order {
  id: UUID
  status: 'pending' | 'confirmed' | 'fulfilled' | 'cancelled'
  items: OrderItem[]
  created_at: string
}
