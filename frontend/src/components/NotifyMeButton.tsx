import { useEffect, useState } from 'react'
import { toast } from 'sonner'
import { isPushSupported, subscribeToPush, getExistingSubscription } from '../push/subscribe'
import { Button } from './ui/button'

type Status = 'checking' | 'unsupported' | 'blocked' | 'subscribed' | 'available'

export default function NotifyMeButton() {
  const [status, setStatus] = useState<Status>('checking')

  useEffect(() => {
    let cancelled = false
    async function check() {
      if (!isPushSupported()) {
        if (!cancelled) setStatus('unsupported')
        return
      }
      if (Notification.permission === 'denied') {
        if (!cancelled) setStatus('blocked')
        return
      }
      const existing = await getExistingSubscription()
      if (!cancelled) setStatus(existing ? 'subscribed' : 'available')
    }
    check()
    return () => {
      cancelled = true
    }
  }, [])

  async function handleClick() {
    try {
      await subscribeToPush()
      setStatus('subscribed')
    } catch {
      if (Notification.permission === 'denied') {
        setStatus('blocked')
      } else {
        toast.error('Could not enable notifications')
      }
    }
  }

  if (status === 'checking' || status === 'unsupported') return null

  if (status === 'blocked') {
    return <p className="text-xs text-muted-foreground">Notifications blocked — enable in browser settings</p>
  }

  if (status === 'subscribed') {
    return <p className="text-xs text-muted-foreground">Notifications on</p>
  }

  return (
    <Button variant="outline" size="sm" onClick={handleClick}>
      Get notified about this order
    </Button>
  )
}
