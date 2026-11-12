import { subscribePush } from '../api/push'

function urlBase64ToUint8Array(base64String: string): Uint8Array<ArrayBuffer> {
  const padding = '='.repeat((4 - (base64String.length % 4)) % 4)
  const base64 = (base64String + padding).replace(/-/g, '+').replace(/_/g, '/')
  const rawData = window.atob(base64)
  const outputArray = new Uint8Array(rawData.length)
  for (let i = 0; i < rawData.length; i++) {
    outputArray[i] = rawData.charCodeAt(i)
  }
  return outputArray
}

const PERMISSION_ASKED_KEY = 'push-permission-asked'

export function isPushSupported(): boolean {
  return 'serviceWorker' in navigator && 'PushManager' in window
}

async function ensureSubscription(): Promise<PushSubscription> {
  const registration = await navigator.serviceWorker.register('/sw.js')
  const existing = await registration.pushManager.getSubscription()
  const subscription =
    existing ??
    (await registration.pushManager.subscribe({
      userVisibleOnly: true,
      applicationServerKey: urlBase64ToUint8Array(import.meta.env.VITE_VAPID_PUBLIC_KEY as string),
    }))
  const json = subscription.toJSON()
  await subscribePush({
    endpoint: json.endpoint!,
    keys: { p256dh: json.keys!.p256dh, auth: json.keys!.auth },
  })
  return subscription
}

// Asks for notification permission once, ever (tracked via localStorage —
// re-prompting on every load gets denied by users on reflex). Subscribes
// immediately whenever permission already allows it, so the backend can
// push at any time without further user action.
export async function initPush(): Promise<void> {
  if (!isPushSupported()) return

  let permission = Notification.permission
  if (permission === 'default' && !localStorage.getItem(PERMISSION_ASKED_KEY)) {
    localStorage.setItem(PERMISSION_ASKED_KEY, '1')
    permission = await Notification.requestPermission()
  }

  if (permission === 'granted') {
    await ensureSubscription()
  }
}
