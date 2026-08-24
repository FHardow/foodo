import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { apiFetch } from './client'

vi.mock('../auth/keycloak', () => ({
  default: {
    updateToken: vi.fn().mockResolvedValue(true),
    token: 'mock-token',
    login: vi.fn(),
  },
}))

describe('apiFetch', () => {
  beforeEach(() => {
    vi.stubGlobal('fetch', vi.fn())
  })
  afterEach(() => vi.unstubAllGlobals())

  it('returns parsed JSON on 200', async () => {
    vi.mocked(fetch).mockResolvedValue(
      new Response(JSON.stringify({ id: '1' }), { status: 200 })
    )
    const result = await apiFetch('/test')
    expect(result).toEqual({ id: '1' })
  })

  it('throws on non-2xx response', async () => {
    vi.mocked(fetch).mockResolvedValue(
      new Response('Not Found', { status: 404 })
    )
    await expect(apiFetch('/test')).rejects.toThrow('404')
  })

  it('resolves without throwing on 204 No Content', async () => {
    vi.mocked(fetch).mockResolvedValue(
      new Response(null, { status: 204 })
    )
    await expect(apiFetch('/test')).resolves.toBeUndefined()
  })
})
