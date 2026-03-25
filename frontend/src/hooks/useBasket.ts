import { useEffect } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { useBasketStore } from '../store/basket'
import * as ordersApi from '../api/orders'
import type { Product } from '../types'

export function useBasket() {
  const {
    basketOrderId,
    isValidating,
    setBasketOrderId,
    clearBasketOrderId,
    setValidating,
  } = useBasketStore()
  const queryClient = useQueryClient()

  // Validate stored order on mount
  useEffect(() => {
    if (!basketOrderId) return
    setValidating(true)
    ordersApi
      .getOrder(basketOrderId)
      .then((order) => {
        if (order.status !== 'pending') clearBasketOrderId()
      })
      .catch(() => clearBasketOrderId())
      .finally(() => setValidating(false))
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  async function addItem(product: Product) {
    try {
      let orderId = basketOrderId
      if (!orderId) {
        const order = await ordersApi.createOrder()
        setBasketOrderId(order.id)
        orderId = order.id
      }
      await ordersApi.addItem(orderId, product.id, 1)
      queryClient.invalidateQueries({ queryKey: ['order', orderId] })
    } catch {
      toast.error('Something went wrong, try again')
    }
  }

  async function removeItem(productId: string) {
    if (!basketOrderId) return
    try {
      await ordersApi.removeItem(basketOrderId, productId)
      queryClient.invalidateQueries({ queryKey: ['order', basketOrderId] })
    } catch {
      toast.error('Something went wrong, try again')
    }
  }

  async function updateQuantity(productId: string, newQuantity: number) {
    if (!basketOrderId) return
    try {
      await ordersApi.removeItem(basketOrderId, productId)
      await ordersApi.addItem(basketOrderId, productId, newQuantity)
      queryClient.invalidateQueries({ queryKey: ['order', basketOrderId] })
    } catch {
      toast.error('Something went wrong, try again')
    }
  }

  async function confirm(): Promise<string | null> {
    if (!basketOrderId) return null
    const order = await ordersApi.confirmOrder(basketOrderId)
    const confirmedId = order.id
    clearBasketOrderId()
    return confirmedId
  }

  return { basketOrderId, isValidating, addItem, removeItem, updateQuantity, confirm }
}
