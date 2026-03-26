import keycloak from '../auth/keycloak'
import { apiFetch } from './client'
import type { Product } from '../types'

const BASE_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'

export const getProducts = () =>
  apiFetch<Product[]>('/api/v1/products')

export interface ProductInput {
  name: string
  description: string
  unit: string
  available: boolean
  price_cents?: number
}

export const createProduct = (data: ProductInput) =>
  apiFetch<Product>('/api/v1/products', {
    method: 'POST',
    body: JSON.stringify({ price_cents: 0, ...data }),
  })

export const updateProduct = (id: string, data: ProductInput) =>
  apiFetch<Product>(`/api/v1/products/${id}`, {
    method: 'PUT',
    body: JSON.stringify({ price_cents: 0, ...data }),
  })

export const deleteProduct = (id: string) =>
  apiFetch<void>(`/api/v1/products/${id}`, { method: 'DELETE' })

export const setAvailability = (id: string, available: boolean) =>
  apiFetch<Product>(`/api/v1/products/${id}/availability`, {
    method: 'PATCH',
    body: JSON.stringify({ available }),
  })

// uploadImage uses a raw fetch because apiFetch forces Content-Type: application/json,
// which breaks multipart/form-data uploads.
export const uploadImage = async (id: string, file: File): Promise<Product> => {
  await keycloak.updateToken(30).catch(() => keycloak.login())
  const form = new FormData()
  form.append('image', file)
  const res = await fetch(`${BASE_URL}/api/v1/products/${id}/image`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${keycloak.token}` },
    body: form,
  })
  if (!res.ok) throw new Error(`${res.status} ${res.statusText}`)
  return res.json() as Promise<Product>
}
