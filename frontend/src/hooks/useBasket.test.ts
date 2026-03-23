import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook, waitFor, act } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import React from 'react'
import { useBasket } from './useBasket'
import { useBasketStore } from '../store/basket'
import * as ordersApi from '../api/orders'
import { toast } from 'sonner'

vi.mock('../api/orders')
vi.mock('sonner')

function makeWrapper() {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: qc }, children)
}

const pendingOrder = (id = 'order-1') => ({
  id, status: 'pending' as const, items: [], created_at: new Date().toISOString(),
})

beforeEach(() => {
  useBasketStore.setState({ basketOrderId: null, isValidating: false })
  localStorage.clear()
  vi.clearAllMocks()
})

describe('useBasket — init validation', () => {
  it('clears stale basket id on 404', async () => {
    useBasketStore.setState({ basketOrderId: 'stale-id' })
    vi.mocked(ordersApi.getOrder).mockRejectedValue(new Error('404'))

    renderHook(() => useBasket(), { wrapper: makeWrapper() })

    await waitFor(() => {
      expect(useBasketStore.getState().basketOrderId).toBeNull()
    })
  })

  it('clears basket id if order is not pending', async () => {
    useBasketStore.setState({ basketOrderId: 'confirmed-id' })
    vi.mocked(ordersApi.getOrder).mockResolvedValue({
      id: 'confirmed-id', status: 'confirmed', items: [], created_at: '',
    })

    renderHook(() => useBasket(), { wrapper: makeWrapper() })

    await waitFor(() => {
      expect(useBasketStore.getState().basketOrderId).toBeNull()
    })
  })

  it('keeps valid pending basket id', async () => {
    useBasketStore.setState({ basketOrderId: 'good-id' })
    vi.mocked(ordersApi.getOrder).mockResolvedValue(pendingOrder('good-id'))

    renderHook(() => useBasket(), { wrapper: makeWrapper() })

    await waitFor(() => {
      expect(useBasketStore.getState().isValidating).toBe(false)
    })
    expect(useBasketStore.getState().basketOrderId).toBe('good-id')
  })
})

describe('useBasket — addItem', () => {
  it('creates order if no basket exists, then adds item', async () => {
    vi.mocked(ordersApi.createOrder).mockResolvedValue(pendingOrder('new-order'))
    vi.mocked(ordersApi.addItem).mockResolvedValue(pendingOrder('new-order'))

    const { result } = renderHook(() => useBasket(), { wrapper: makeWrapper() })

    await act(async () => {
      await result.current.addItem({
        id: 'prod-1', name: 'Bread', description: '', unit: 'loaf', available: true,
      })
    })

    expect(ordersApi.createOrder).toHaveBeenCalledOnce()
    expect(ordersApi.addItem).toHaveBeenCalledWith('new-order', 'prod-1', 1)
    expect(useBasketStore.getState().basketOrderId).toBe('new-order')
  })

  it('reuses existing basket order', async () => {
    useBasketStore.setState({ basketOrderId: 'existing-order' })
    vi.mocked(ordersApi.getOrder).mockResolvedValue(pendingOrder('existing-order'))
    vi.mocked(ordersApi.addItem).mockResolvedValue(pendingOrder('existing-order'))

    const { result } = renderHook(() => useBasket(), { wrapper: makeWrapper() })
    await waitFor(() => expect(useBasketStore.getState().isValidating).toBe(false))

    await act(async () => {
      await result.current.addItem({
        id: 'prod-2', name: 'Rye', description: '', unit: 'loaf', available: true,
      })
    })

    expect(ordersApi.createOrder).not.toHaveBeenCalled()
    expect(ordersApi.addItem).toHaveBeenCalledWith('existing-order', 'prod-2', 1)
  })
})

describe('useBasket — confirm', () => {
  it('confirms order, clears basket id, returns order id', async () => {
    useBasketStore.setState({ basketOrderId: 'order-to-confirm' })
    vi.mocked(ordersApi.getOrder).mockResolvedValue(pendingOrder('order-to-confirm'))
    vi.mocked(ordersApi.confirmOrder).mockResolvedValue({
      id: 'order-to-confirm', status: 'confirmed', items: [], created_at: '',
    })

    const { result } = renderHook(() => useBasket(), { wrapper: makeWrapper() })
    await waitFor(() => expect(useBasketStore.getState().isValidating).toBe(false))

    let returnedId: string | null = null
    await act(async () => { returnedId = await result.current.confirm() })

    expect(returnedId).toBe('order-to-confirm')
    expect(useBasketStore.getState().basketOrderId).toBeNull()
  })

  it('shows toast on confirm failure, does not clear basket id', async () => {
    useBasketStore.setState({ basketOrderId: 'order-to-confirm' })
    vi.mocked(ordersApi.getOrder).mockResolvedValue(pendingOrder('order-to-confirm'))
    vi.mocked(ordersApi.confirmOrder).mockRejectedValue(new Error('500'))

    const { result } = renderHook(() => useBasket(), { wrapper: makeWrapper() })
    await waitFor(() => expect(useBasketStore.getState().isValidating).toBe(false))

    let returnedId: string | null = 'sentinel'
    await act(async () => { returnedId = await result.current.confirm() })

    expect(returnedId).toBeNull()
    expect(useBasketStore.getState().basketOrderId).toBe('order-to-confirm')
    expect(toast.error).toHaveBeenCalledWith('Something went wrong, try again')
  })
})
