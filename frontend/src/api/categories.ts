import { apiFetch } from './client'
import type { Category } from '../types'

export const getCategories = () => apiFetch<Category[]>('/api/v1/categories')

export const deleteCategory = (id: string) =>
  apiFetch<void>(`/api/v1/categories/${id}`, { method: 'DELETE' })
