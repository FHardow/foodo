import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { apiFetch } from './client'

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
})
