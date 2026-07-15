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
import { Button } from '../../components/ui/button'
import { Input } from '../../components/ui/input'
import { Label } from '../../components/ui/label'
import { Textarea } from '../../components/ui/textarea'
import { Switch } from '../../components/ui/switch'
import { Dialog, DialogContent, DialogHeader, DialogFooter, DialogTitle } from '../../components/ui/dialog'

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
    return <div className="text-center py-16 text-muted-foreground">Loading…</div>
  }

  if (isError) {
    return <div className="text-center py-16 text-muted-foreground">Failed to load products.</div>
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold text-primary">Products</h1>
        <Button onClick={openCreate}>+ Add product</Button>
      </div>

      {products.length === 0 ? (
        <p className="text-muted-foreground text-center py-16">No products yet.</p>
      ) : (
        <div className="bg-card rounded-lg border border-border overflow-hidden">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-border text-left text-muted-foreground">
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
                <tr key={p.id} className="border-b border-border last:border-0">
                  <td className="px-4 py-3">
                    {p.image_url ? (
                      <img
                        src={`${BASE_URL}${p.image_url}`}
                        alt={p.name}
                        className="w-10 h-10 object-cover rounded"
                      />
                    ) : (
                      <div className="w-10 h-10 bg-secondary rounded flex items-center justify-center text-muted-foreground text-xs">
                        No img
                      </div>
                    )}
                  </td>
                  <td className="px-4 py-3 font-medium text-foreground">{p.name}</td>
                  <td className="px-4 py-3 hidden sm:table-cell text-muted-foreground">{p.unit}</td>
                  <td className="px-4 py-3 hidden md:table-cell text-muted-foreground max-w-xs truncate">
                    {p.description}
                  </td>
                  <td className="px-4 py-3">
                    <Switch
                      checked={p.available}
                      onCheckedChange={(checked) =>
                        availabilityMutation.mutate({ id: p.id, available: checked })
                      }
                      aria-label={p.available ? 'Mark unavailable' : 'Mark available'}
                    />
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-2 justify-end">
                      {deleteConfirm === p.id ? (
                        <>
                          <span className="text-muted-foreground text-xs">Delete?</span>
                          <Button
                            variant="link"
                            className="text-red-600 text-xs h-auto p-0"
                            onClick={() => deleteMutation.mutate(p.id)}
                          >
                            Yes
                          </Button>
                          <Button
                            variant="link"
                            className="text-muted-foreground text-xs h-auto p-0"
                            onClick={() => setDeleteConfirm(null)}
                          >
                            No
                          </Button>
                        </>
                      ) : (
                        <>
                          <Button
                            variant="link"
                            className="text-primary text-xs h-auto p-0"
                            onClick={() => openEdit(p)}
                          >
                            Edit
                          </Button>
                          <Button
                            variant="link"
                            className="text-red-500 text-xs h-auto p-0"
                            onClick={() => setDeleteConfirm(p.id)}
                          >
                            Delete
                          </Button>
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

      <Dialog open={modalOpen} onOpenChange={setModalOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editing ? 'Edit product' : 'Add product'}</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleSubmit} className="px-6 py-4 space-y-4">
            <div className="space-y-1">
              <Label htmlFor="product-name">
                Name <span className="text-red-400">*</span>
              </Label>
              <Input
                id="product-name"
                required
                value={form.name}
                onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
              />
            </div>
            <div className="space-y-1">
              <Label htmlFor="product-unit">
                Unit <span className="text-red-400">*</span>
              </Label>
              <Input
                id="product-unit"
                required
                placeholder="e.g. loaf, dozen"
                value={form.unit}
                onChange={(e) => setForm((f) => ({ ...f, unit: e.target.value }))}
              />
            </div>
            <div className="space-y-1">
              <Label htmlFor="product-description">Description</Label>
              <Textarea
                id="product-description"
                rows={3}
                value={form.description}
                onChange={(e) => setForm((f) => ({ ...f, description: e.target.value }))}
              />
            </div>
            <div className="flex items-center gap-2">
              <Switch
                id="available"
                checked={form.available}
                onCheckedChange={(checked) => setForm((f) => ({ ...f, available: checked }))}
              />
              <Label htmlFor="available">Available for ordering</Label>
            </div>
            <div>
              <Label className="block mb-1">Image</Label>
              {editing?.image_url && !imageFile && (
                <img
                  src={`${BASE_URL}${editing.image_url}`}
                  alt="current"
                  className="w-16 h-16 object-cover rounded mb-2"
                />
              )}
              <Label
                htmlFor="product-image"
                className="inline-flex items-center gap-2 cursor-pointer border border-primary text-primary rounded px-3 py-1.5 text-sm hover:bg-primary hover:text-primary-foreground transition-colors font-normal"
              >
                {imageFile ? imageFile.name : 'Choose image…'}
                <input
                  id="product-image"
                  type="file"
                  accept="image/*"
                  onChange={(e) => setImageFile(e.target.files?.[0] ?? null)}
                  className="sr-only"
                />
              </Label>
              {imageFile && (
                <Button
                  type="button"
                  variant="link"
                  className="ml-2 text-xs text-muted-foreground hover:text-red-500 h-auto p-0"
                  onClick={() => setImageFile(null)}
                >
                  Remove
                </Button>
              )}
            </div>
            <DialogFooter className="px-0 pb-0">
              <Button type="button" variant="ghost" onClick={() => setModalOpen(false)}>
                Cancel
              </Button>
              <Button type="submit" disabled={submitting}>
                {submitting ? 'Saving…' : editing ? 'Save changes' : 'Create product'}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  )
}
