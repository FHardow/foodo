import { describe, it, expect, beforeEach } from 'vitest'
import { useBasketStore } from './basket'

beforeEach(() => {
  useBasketStore.setState({ basketOrderId: null, isValidating: false })
  localStorage.clear()
})

describe('basket store', () => {
  it('starts with null basketOrderId', () => {
    expect(useBasketStore.getState().basketOrderId).toBeNull()
  })

  it('setBasketOrderId stores the id', () => {
    useBasketStore.getState().setBasketOrderId('abc-123')
    expect(useBasketStore.getState().basketOrderId).toBe('abc-123')
  })

  it('clearBasketOrderId resets to null', () => {
    useBasketStore.getState().setBasketOrderId('abc-123')
    useBasketStore.getState().clearBasketOrderId()
    expect(useBasketStore.getState().basketOrderId).toBeNull()
  })

  it('setValidating toggles isValidating', () => {
    useBasketStore.getState().setValidating(true)
    expect(useBasketStore.getState().isValidating).toBe(true)
    useBasketStore.getState().setValidating(false)
    expect(useBasketStore.getState().isValidating).toBe(false)
  })

  it('persists basketOrderId to localStorage', () => {
    useBasketStore.getState().setBasketOrderId('persist-me')
    const raw = localStorage.getItem('basketOrderId')
    expect(raw).not.toBeNull()
    const stored = JSON.parse(raw!)
    expect(stored.state.basketOrderId).toBe('persist-me')
  })

  it('does not persist isValidating', () => {
    useBasketStore.getState().setValidating(true)
    const raw = localStorage.getItem('basketOrderId')
    const stored = raw ? JSON.parse(raw) : {}
    expect(stored?.state?.isValidating).toBeUndefined()
  })
})
