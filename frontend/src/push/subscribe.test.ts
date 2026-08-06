import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { isPushSupported, subscribeToPush } from './subscribe'
import * as pushApi from '../api/push'

vi.mock('../api/push')

beforeEach(() => {
  // subscribeToPush() reads this at call time (not at module load), so
  // stubbing it fresh per test is sufficient — no dynamic re-import needed.
  vi.stubEnv('VITE_VAPID_PUBLIC_KEY', 'BEl62iUYgUivxIkv69yViEuiBIa40HI80NM9-_uk_Vd0mrs4X6qKZfDgO2c5NEnXwuh3AteEBzY-Ov6uJU')
})

afterEach(() => {
  vi.unstubAllGlobals()
  vi.unstubAllEnvs()
  vi.clearAllMocks()
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

describe('subscribeToPush', () => {
  const fakeSubscription = {
    endpoint: 'https://push.example.com/abc',
    toJSON: () => ({
      endpoint: 'https://push.example.com/abc',
      keys: { p256dh: 'p256dh-key', auth: 'auth-key' },
    }),
  }

  beforeEach(() => {
    vi.stubGlobal('Notification', { requestPermission: vi.fn().mockResolvedValue('granted') })
    Object.defineProperty(navigator, 'serviceWorker', {
      value: {
        register: vi.fn().mockResolvedValue({
          pushManager: { subscribe: vi.fn().mockResolvedValue(fakeSubscription) },
        }),
      },
      configurable: true,
    })
  })

  it('registers the service worker, subscribes, and posts to the backend', async () => {
    await subscribeToPush()
    expect(pushApi.subscribePush).toHaveBeenCalledWith({
      endpoint: 'https://push.example.com/abc',
      keys: { p256dh: 'p256dh-key', auth: 'auth-key' },
    })
  })

  it('throws when permission is denied', async () => {
    vi.stubGlobal('Notification', { requestPermission: vi.fn().mockResolvedValue('denied') })
    await expect(subscribeToPush()).rejects.toThrow('denied')
  })
})
