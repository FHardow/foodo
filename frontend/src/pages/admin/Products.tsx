import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import keycloak from '../../auth/keycloak'
import {
  getProducts,
  createProduct,
  updateProduct,
  deleteProduct,
  setAvailability,
  uploadImage,
} from '../../api/products'
import type { ProductInput } from '../../api/products'
import type { Product } from '../../types'

const BASE_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'

const emptyForm: ProductInput = { name: '', description: '', unit: '', available: false }

export default function AdminProducts() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()

  const { data: products = [], isLoading, isError } = useQuery({
    queryKey: ['products'],
    queryFn: getProducts,
  })

  const [modalOpen, setModalOpen] = useState(false)
  const [editing, setEditing] = useState<Product | null>(null)
  const [form, setForm] = useState<ProductInput>(emptyForm)
  const [imageFile, setImageFile] = useState<File | null>(null)
  const [deleteConfirm, setDeleteConfirm] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  const invalidate = () => queryClient.invalidateQueries({ queryKey: ['products'] })

  const availabilityMutation = useMutation({
    mutationFn: ({ id, available }: { id: string; available: boolean }) =>
      setAvailability(id, available),
    onSuccess: invalidate,
    onError: () => toast.error('Failed to update availability'),
  })

  const deleteMutation = useMutation({
    mutationFn: deleteProduct,
    onSuccess: () => {
      invalidate()
      setDeleteConfirm(null)
      toast.success('Product deleted')
    },
    onError: () => toast.error('Failed to delete product'),
  })

  if (!keycloak.hasRealmRole('owner')) {
    navigate('/')
    return null
  }

  const openCreate = () => {
    setEditing(null)
    setForm(emptyForm)
    setImageFile(null)
    setModalOpen(true)
  }

  const openEdit = (p: Product) => {
    setEditing(p)
    setForm({ name: p.name, description: p.description, unit: p.unit, available: p.available })
    setImageFile(null)
    setModalOpen(true)
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setSubmitting(true)
    try {
      let product: Product
      if (editing) {
        product = await updateProduct(editing.id, form)
        toast.success('Product updated')
      } else {
        product = await createProduct(form)
        toast.success('Product created')
      }
      if (imageFile) {
        await uploadImage(product.id, imageFile)
      }
      invalidate()
      setModalOpen(false)
    } catch {
      toast.error('Something went wrong')
    } finally {
      setSubmitting(false)
    }
  }

  if (isLoading) {
    return <div className="text-center py-16 text-[#8a6a50]">Loading…</div>
  }

  if (isError) {
    return <div className="text-center py-16 text-[#8a6a50]">Failed to load products.</div>
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold text-[#5c3d1e]">Products</h1>
        <button
          onClick={openCreate}
          className="bg-[#5c3d1e] text-white rounded px-4 py-2 text-sm hover:bg-[#3d2b1a] transition-colors"
        >
          + Add product
        </button>
      </div>

      {products.length === 0 ? (
        <p className="text-[#8a6a50] text-center py-16">No products yet.</p>
      ) : (
        <div className="bg-white rounded-lg border border-[#e8ddd0] overflow-hidden">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-[#e8ddd0] text-left text-[#8a6a50]">
                <th className="px-4 py-3 w-12"></th>
                <th className="px-4 py-3">Name</th>
                <th className="px-4 py-3 hidden sm:table-cell">Unit</th>
                <th className="px-4 py-3 hidden md:table-cell">Description</th>
                <th className="px-4 py-3">Available</th>
                <th className="px-4 py-3"></th>
              </tr>
            </thead>
            <tbody>
              {products.map((p) => (
                <tr key={p.id} className="border-b border-[#e8ddd0] last:border-0">
                  <td className="px-4 py-3">
                    {p.image_url ? (
                      <img
                        src={`${BASE_URL}${p.image_url}`}
                        alt={p.name}
                        className="w-10 h-10 object-cover rounded"
                      />
                    ) : (
                      <div className="w-10 h-10 bg-[#f0e8de] rounded flex items-center justify-center text-[#c4a882] text-xs">
                        No img
                      </div>
                    )}
                  </td>
                  <td className="px-4 py-3 font-medium text-[#3d2b1a]">{p.name}</td>
                  <td className="px-4 py-3 hidden sm:table-cell text-[#8a6a50]">{p.unit}</td>
                  <td className="px-4 py-3 hidden md:table-cell text-[#8a6a50] max-w-xs truncate">
                    {p.description}
                  </td>
                  <td className="px-4 py-3">
                    <button
                      onClick={() => availabilityMutation.mutate({ id: p.id, available: !p.available })}
                      className={`w-10 h-6 rounded-full transition-colors ${
                        p.available ? 'bg-[#5c3d1e]' : 'bg-[#e8ddd0]'
                      }`}
                      aria-label={p.available ? 'Mark unavailable' : 'Mark available'}
                    >
                      <span
                        className={`block w-4 h-4 bg-white rounded-full mx-1 transition-transform ${
                          p.available ? 'translate-x-4' : 'translate-x-0'
                        }`}
                      />
                    </button>
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-2 justify-end">
                      {deleteConfirm === p.id ? (
                        <>
                          <span className="text-[#8a6a50] text-xs">Delete?</span>
                          <button
                            onClick={() => deleteMutation.mutate(p.id)}
                            className="text-red-600 text-xs font-medium hover:underline"
                          >
                            Yes
                          </button>
                          <button
                            onClick={() => setDeleteConfirm(null)}
                            className="text-[#8a6a50] text-xs hover:underline"
                          >
                            No
                          </button>
                        </>
                      ) : (
                        <>
                          <button
                            onClick={() => openEdit(p)}
                            className="text-[#5c3d1e] text-xs font-medium hover:underline"
                          >
                            Edit
                          </button>
                          <button
                            onClick={() => setDeleteConfirm(p.id)}
                            className="text-red-500 text-xs font-medium hover:underline"
                          >
                            Delete
                          </button>
                        </>
                      )}
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Modal */}
      {modalOpen && (
        <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-lg shadow-xl w-full max-w-md">
            <div className="px-6 py-4 border-b border-[#e8ddd0] flex justify-between items-center">
              <h2 className="font-semibold text-[#3d2b1a]">
                {editing ? 'Edit product' : 'Add product'}
              </h2>
              <button
                onClick={() => setModalOpen(false)}
                className="text-[#8a6a50] hover:text-[#3d2b1a] text-xl leading-none"
              >
                ×
              </button>
            </div>
            <form onSubmit={handleSubmit} className="px-6 py-4 space-y-4">
              <div>
                <label className="block text-xs font-medium text-[#8a6a50] mb-1">
                  Name <span className="text-red-400">*</span>
                </label>
                <input
                  required
                  value={form.name}
                  onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
                  className="w-full border border-[#e8ddd0] rounded px-3 py-2 text-sm focus:outline-none focus:border-[#5c3d1e]"
                />
              </div>
              <div>
                <label className="block text-xs font-medium text-[#8a6a50] mb-1">
                  Unit <span className="text-red-400">*</span>
                </label>
                <input
                  required
                  placeholder="e.g. loaf, dozen"
                  value={form.unit}
                  onChange={(e) => setForm((f) => ({ ...f, unit: e.target.value }))}
                  className="w-full border border-[#e8ddd0] rounded px-3 py-2 text-sm focus:outline-none focus:border-[#5c3d1e]"
                />
              </div>
              <div>
                <label className="block text-xs font-medium text-[#8a6a50] mb-1">Description</label>
                <textarea
                  rows={3}
                  value={form.description}
                  onChange={(e) => setForm((f) => ({ ...f, description: e.target.value }))}
                  className="w-full border border-[#e8ddd0] rounded px-3 py-2 text-sm focus:outline-none focus:border-[#5c3d1e] resize-none"
                />
              </div>
              <div className="flex items-center gap-2">
                <input
                  id="available"
                  type="checkbox"
                  checked={form.available}
                  onChange={(e) => setForm((f) => ({ ...f, available: e.target.checked }))}
                  className="accent-[#5c3d1e]"
                />
                <label htmlFor="available" className="text-sm text-[#3d2b1a]">
                  Available for ordering
                </label>
              </div>
              <div>
                <label className="block text-xs font-medium text-[#8a6a50] mb-1">Image</label>
                {editing?.image_url && !imageFile && (
                  <img
                    src={`${BASE_URL}${editing.image_url}`}
                    alt="current"
                    className="w-16 h-16 object-cover rounded mb-2"
                  />
                )}
                <label className="inline-flex items-center gap-2 cursor-pointer border border-[#5c3d1e] text-[#5c3d1e] rounded px-3 py-1.5 text-sm hover:bg-[#5c3d1e] hover:text-white transition-colors">
                  {imageFile ? imageFile.name : 'Choose image…'}
                  <input
                    type="file"
                    accept="image/*"
                    onChange={(e) => setImageFile(e.target.files?.[0] ?? null)}
                    className="sr-only"
                  />
                </label>
                {imageFile && (
                  <button
                    type="button"
                    onClick={() => setImageFile(null)}
                    className="ml-2 text-xs text-[#8a6a50] hover:text-red-500"
                  >
                    Remove
                  </button>
                )}
              </div>
              <div className="flex justify-end gap-3 pt-2">
                <button
                  type="button"
                  onClick={() => setModalOpen(false)}
                  className="px-4 py-2 text-sm text-[#8a6a50] hover:text-[#3d2b1a]"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={submitting}
                  className="bg-[#5c3d1e] text-white rounded px-4 py-2 text-sm hover:bg-[#3d2b1a] disabled:opacity-50 transition-colors"
                >
                  {submitting ? 'Saving…' : editing ? 'Save changes' : 'Create product'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
