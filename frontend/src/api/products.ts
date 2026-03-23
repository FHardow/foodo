import { apiFetch } from './client'
import type { Product } from '../types'

export const getProducts = () =>
  apiFetch<Product[]>('/api/v1/products')
