import { apiFetch } from './client'

export const subscribePush = (sub: { endpoint: string; keys: { p256dh: string; auth: string } }) =>
  apiFetch('/api/v1/push/subscribe', { method: 'POST', body: JSON.stringify(sub) })

export const unsubscribePush = (endpoint: string) =>
  apiFetch('/api/v1/push/subscribe', { method: 'DELETE', body: JSON.stringify({ endpoint }) })
