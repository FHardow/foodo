import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface BasketState {
  basketOrderId: string | null
  isValidating: boolean
  setBasketOrderId: (id: string) => void
  clearBasketOrderId: () => void
  setValidating: (v: boolean) => void
}

export const useBasketStore = create<BasketState>()(
  persist(
    (set) => ({
      basketOrderId: null,
      isValidating: false,
      setBasketOrderId: (id) => set({ basketOrderId: id }),
      clearBasketOrderId: () => set({ basketOrderId: null }),
      setValidating: (v) => set({ isValidating: v }),
    }),
    {
      name: 'basketOrderId',
      partialize: (s) => ({ basketOrderId: s.basketOrderId }),
    }
  )
)
