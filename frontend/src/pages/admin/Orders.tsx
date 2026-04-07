import { useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import {
  DndContext,
  DragOverlay,
  useDroppable,
  useDraggable,
  type DragEndEvent,
  type DragStartEvent,
  PointerSensor,
  useSensor,
  useSensors,
} from '@dnd-kit/core'
import { CSS } from '@dnd-kit/utilities'
import { useState, useEffect } from 'react'
import keycloak from '../../auth/keycloak'
import { getAllOrders, acceptOrder, startOrder, finishOrder } from '../../api/orders'
import type { Order } from '../../types'

type KanbanStatus = 'created' | 'accepted' | 'ongoing' | 'finished'

const COLUMNS: { status: KanbanStatus; label: string }[] = [
  { status: 'created',  label: 'New Orders' },
  { status: 'accepted', label: 'Accepted' },
  { status: 'ongoing',  label: 'Ongoing' },
  { status: 'finished', label: 'Finished' },
]

const STATUS_ORDER: KanbanStatus[] = ['created', 'accepted', 'ongoing', 'finished']

// ---- Draggable card ----

function OrderCard({ order, isDragging }: { order: Order; isDragging?: boolean }) {
  const { attributes, listeners, setNodeRef, transform } = useDraggable({ id: order.id })

  const style = {
    transform: CSS.Translate.toString(transform),
    opacity: isDragging ? 0.4 : 1,
  }

  return (
    <div
      ref={setNodeRef}
      style={style}
      {...listeners}
      {...attributes}
      data-testid="order-card"
      data-order-id={order.id}
      data-order-status={order.status}
      className="bg-white rounded-lg border border-[#e8ddd0] p-3 cursor-grab active:cursor-grabbing select-none shadow-sm hover:border-[#5c3d1e] transition-colors"
    >
      <CardContent order={order} />
    </div>
  )
}

function CardContent({ order }: { order: Order }) {
  return (
    <>
      <div className="flex items-start justify-between gap-2 mb-1">
        <span className="text-xs font-medium text-[#5c3d1e] truncate">
          {order.user_name ?? 'Unknown customer'}
        </span>
        <span className="text-xs text-[#8a6a50] whitespace-nowrap">
          {new Date(order.created_at).toLocaleDateString()}
        </span>
      </div>
      <ul className="mt-1 space-y-0.5">
        {order.items.map((item) => (
          <li key={item.product_id} className="text-xs text-[#3d2b1a]">
            {item.product_name}
            {item.unit ? ` · ${item.unit} × ${item.quantity}` : ` × ${item.quantity}`}
          </li>
        ))}
      </ul>
    </>
  )
}

// ---- Droppable column ----

function KanbanColumn({
  status,
  label,
  orders,
  activeId,
}: {
  status: KanbanStatus
  label: string
  orders: Order[]
  activeId: string | null
}) {
  const { setNodeRef, isOver } = useDroppable({ id: status })

  return (
    <div
      ref={setNodeRef}
      data-testid="kanban-column"
      data-column-status={status}
      className={`flex flex-col min-h-[200px] rounded-xl p-3 transition-colors ${
        isOver ? 'bg-[#e8ddd0] ring-2 ring-[#5c3d1e]' : 'bg-[#f0e8de]'
      }`}
    >
      <div className="flex items-center justify-between mb-3">
        <h2 className="font-semibold text-sm text-[#3d2b1a]">{label}</h2>
        <span data-testid="column-count" className="text-xs bg-white text-[#8a6a50] rounded-full px-2 py-0.5">
          {orders.length}
        </span>
      </div>
      <div className="flex flex-col gap-2 flex-1">
        {orders.map((order) => (
          <OrderCard key={order.id} order={order} isDragging={order.id === activeId} />
        ))}
        {orders.length === 0 && (
          <p className="text-xs text-[#8a6a50] text-center mt-4">No orders</p>
        )}
      </div>
    </div>
  )
}

// ---- Main page ----

export default function AdminOrders() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [activeId, setActiveId] = useState<string | null>(null)
  const isOwner = keycloak.hasRealmRole('owner')

  const { data: allOrders = [], isLoading, isError } = useQuery({
    queryKey: ['all-orders'],
    queryFn: getAllOrders,
    refetchInterval: 30_000,
    enabled: isOwner,
  })

  const invalidate = () => queryClient.invalidateQueries({ queryKey: ['all-orders'] })

  const accept = useMutation({ mutationFn: acceptOrder, onSuccess: invalidate, onError: () => toast.error('Failed to accept order') })
  const start  = useMutation({ mutationFn: startOrder,  onSuccess: invalidate, onError: () => toast.error('Failed to start order') })
  const finish = useMutation({ mutationFn: finishOrder, onSuccess: invalidate, onError: () => toast.error('Failed to finish order') })

  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 5 } })
  )

  useEffect(() => {
    if (!isOwner) navigate('/')
  }, [isOwner, navigate])

  if (!isOwner) return null

  // Only show non-pending orders in the kanban
  const kanbanOrders = allOrders.filter(
    (o): o is Order & { status: KanbanStatus } => o.status !== 'pending'
  )

  const columnOrders = (status: KanbanStatus) =>
    kanbanOrders.filter((o) => o.status === status)

  const activeOrder = activeId ? kanbanOrders.find((o) => o.id === activeId) : null

  function handleDragStart({ active }: DragStartEvent) {
    setActiveId(active.id as string)
  }

  function handleDragEnd({ active, over }: DragEndEvent) {
    setActiveId(null)
    if (!over) return

    const fromOrder = kanbanOrders.find((o) => o.id === active.id)
    if (!fromOrder) return

    const fromIdx = STATUS_ORDER.indexOf(fromOrder.status as KanbanStatus)
    const toIdx   = STATUS_ORDER.indexOf(over.id as KanbanStatus)

    if (toIdx !== fromIdx + 1) return // only forward, one step at a time

    const orderId = active.id as string
    if (toIdx === 1) accept.mutate(orderId)
    else if (toIdx === 2) start.mutate(orderId)
    else if (toIdx === 3) finish.mutate(orderId)
  }

  if (isLoading) {
    return <div data-testid="loading-skeleton" className="animate-pulse bg-white rounded-lg h-48 border border-[#e8ddd0]" />
  }

  if (isError) {
    return (
      <div className="text-center py-16">
        <p className="text-[#8a6a50]">Could not load orders. Try again.</p>
      </div>
    )
  }

  return (
    <div>
      <h1 className="text-2xl font-bold text-[#3d2b1a] mb-6">Order Board</h1>
      <p className="text-sm text-[#8a6a50] mb-6">
        Drag orders to the next column to advance their status.
      </p>

      <DndContext sensors={sensors} onDragStart={handleDragStart} onDragEnd={handleDragEnd}>
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          {COLUMNS.map(({ status, label }) => (
            <KanbanColumn
              key={status}
              status={status}
              label={label}
              orders={columnOrders(status)}
              activeId={activeId}
            />
          ))}
        </div>

        <DragOverlay>
          {activeOrder && (
            <div className="bg-white rounded-lg border-2 border-[#5c3d1e] p-3 shadow-lg rotate-1 w-64">
              <CardContent order={activeOrder} />
            </div>
          )}
        </DragOverlay>
      </DndContext>
    </div>
  )
}
