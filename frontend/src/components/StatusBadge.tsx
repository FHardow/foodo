import type { Order } from '../types'
import { Badge } from './ui/badge'

const colours: Record<Order['status'], string> = {
  pending:  'bg-amber-100 text-amber-800',
  created:  'bg-blue-100 text-blue-800',
  accepted: 'bg-purple-100 text-purple-800',
  ongoing:  'bg-orange-100 text-orange-800',
  finished: 'bg-green-100 text-green-800',
  archived: 'bg-gray-100 text-gray-600',
}

export default function StatusBadge({ status }: { status: Order['status'] }) {
  return <Badge className={colours[status]}>{status}</Badge>
}
