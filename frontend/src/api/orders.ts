import { apiFetch } from './client'
import keycloak from '../auth/keycloak'
import type { Order } from '../types'

export const createOrder = () =>
  apiFetch<Order>('/api/v1/orders', {
    method: 'POST',
    body: JSON.stringify({ user_id: keycloak.subject }),
  })

export const getOrder = (id: string) =>
  apiFetch<Order>(`/api/v1/orders/${id}`)

export const getOrders = () =>
  apiFetch<Order[]>(`/api/v1/orders?user_id=${keycloak.subject}`)

export const addItem = (orderId: string, productId: string, quantity: number) =>
  apiFetch<Order>(`/api/v1/orders/${orderId}/items`, {
    method: 'POST',
    body: JSON.stringify({ product_id: productId, quantity }),
  })

export const removeItem = (orderId: string, productId: string) =>
  apiFetch<Order>(`/api/v1/orders/${orderId}/items/${productId}`, {
    method: 'DELETE',
  })

export const confirmOrder = (orderId: string) =>
  apiFetch<Order>(`/api/v1/orders/${orderId}/confirm`, { method: 'POST' })

export const getAllOrders = () =>
  apiFetch<Order[]>('/api/v1/orders')

export const acceptOrder = (orderId: string) =>
  apiFetch<Order>(`/api/v1/orders/${orderId}/accept`, { method: 'POST' })

export const startOrder = (orderId: string) =>
  apiFetch<Order>(`/api/v1/orders/${orderId}/start`, { method: 'POST' })

export const finishOrder = (orderId: string) =>
  apiFetch<Order>(`/api/v1/orders/${orderId}/finish`, { method: 'POST' })

export const unacceptOrder = (orderId: string) =>
  apiFetch<Order>(`/api/v1/orders/${orderId}/unaccept`, { method: 'POST' })

export const stopOrder = (orderId: string) =>
  apiFetch<Order>(`/api/v1/orders/${orderId}/stop`, { method: 'POST' })

export const unfinishOrder = (orderId: string) =>
  apiFetch<Order>(`/api/v1/orders/${orderId}/unfinish`, { method: 'POST' })
