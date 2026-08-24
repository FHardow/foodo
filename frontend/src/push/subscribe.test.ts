import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { isPushSupported, initPush } from './subscribe'
import * as pushApi from '../api/push'

vi.mock('../api/push')

beforeEach(() => {
  localStorage.clear()
  // ensureSubscription() reads this at call time (not at module load), so
  // stubbing it fresh per test is sufficient — no dynamic re-import needed.
  vi.stubEnv('VITE_VAPID_PUBLIC_KEY', 'BEl62iUYgUivxIkv69yViEuiBIa40HI80NM9-_uk_Vd0mrs4X6qKZfDgO2c5NEnXwuh3AteEBzY-Ov6uJU')
})

afterEach(() => {
  vi.unstubAllGlobals()
  vi.unstubAllEnvs()
  vi.clearAllMocks()
  localStorage.clear()
})

describe('isPushSupported', () => {
  it('returns false when serviceWorker/PushManager are absent', () => {
    expect(isPushSupported()).toBe(false)
  })

  it('returns true when both are present', () => {
    vi.stubGlobal('PushManager', function () {})
    Object.defineProperty(navigator, 'serviceWorker', { value: {}, configurable: true })
    expect(isPushSupported()).toBe(true)
  })
})

describe('initPush', () => {
  const fakeSubscription = {
    endpoint: 'https://push.example.com/abc',
    toJSON: () => ({
      endpoint: 'https://push.example.com/abc',
      keys: { p256dh: 'p256dh-key', auth: 'auth-key' },
    }),
  }

  function stubServiceWorker(getSubscription: () => Promise<unknown>) {
    Object.defineProperty(navigator, 'serviceWorker', {
      value: {
        register: vi.fn().mockResolvedValue({
          pushManager: {
            getSubscription,
            subscribe: vi.fn().mockResolvedValue(fakeSubscription),
          },
        }),
      },
      configurable: true,
    })
  }

  it('does nothing when push is unsupported', async () => {
    await initPush()
    expect(pushApi.subscribePush).not.toHaveBeenCalled()
  })

  it('requests permission once, then subscribes and posts to the backend', async () => {
    vi.stubGlobal('PushManager', function () {})
    vi.stubGlobal('Notification', { permission: 'default', requestPermission: vi.fn().mockResolvedValue('granted') })
    stubServiceWorker(() => Promise.resolve(null))

    await initPush()

    expect(Notification.requestPermission).toHaveBeenCalledTimes(1)
    expect(localStorage.getItem('push-permission-asked')).toBe('1')
    expect(pushApi.subscribePush).toHaveBeenCalledWith({
      endpoint: 'https://push.example.com/abc',
      keys: { p256dh: 'p256dh-key', auth: 'auth-key' },
    })
  })

  it('does not re-prompt on a later call once permission was already asked', async () => {
    localStorage.setItem('push-permission-asked', '1')
    vi.stubGlobal('PushManager', function () {})
    vi.stubGlobal('Notification', { permission: 'default', requestPermission: vi.fn() })
    stubServiceWorker(() => Promise.resolve(null))

    await initPush()

    expect(Notification.requestPermission).not.toHaveBeenCalled()
    expect(pushApi.subscribePush).not.toHaveBeenCalled()
  })

  it('reuses an existing subscription instead of creating a new one', async () => {
    vi.stubGlobal('PushManager', function () {})
    vi.stubGlobal('Notification', { permission: 'granted', requestPermission: vi.fn() })
    const subscribe = vi.fn()
    Object.defineProperty(navigator, 'serviceWorker', {
      value: {
        register: vi.fn().mockResolvedValue({
          pushManager: { getSubscription: () => Promise.resolve(fakeSubscription), subscribe },
        }),
      },
      configurable: true,
    })

    await initPush()

    expect(subscribe).not.toHaveBeenCalled()
    expect(pushApi.subscribePush).toHaveBeenCalledWith({
      endpoint: 'https://push.example.com/abc',
      keys: { p256dh: 'p256dh-key', auth: 'auth-key' },
    })
  })

  it('does not subscribe when permission is denied', async () => {
    vi.stubGlobal('PushManager', function () {})
    vi.stubGlobal('Notification', { permission: 'denied', requestPermission: vi.fn() })
    stubServiceWorker(() => Promise.resolve(null))

    await initPush()

    expect(pushApi.subscribePush).not.toHaveBeenCalled()
  })
})
