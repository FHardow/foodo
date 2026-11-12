import keycloak from '../auth/keycloak'

const BASE_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'

export async function apiFetch<T = unknown>(path: string, options?: RequestInit): Promise<T> {
  await keycloak.updateToken(30).catch(() => keycloak.login())

  const res = await fetch(`${BASE_URL}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${keycloak.token}`,
      ...options?.headers,
    },
  })
  if (!res.ok) {
    throw new Error(`${res.status} ${res.statusText}`)
  }
  // DELETE endpoints return 204 with an empty body.
  if (res.status === 204) return undefined as T
  return res.json() as Promise<T>
}
