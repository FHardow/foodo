import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import NotifyMeButton from './NotifyMeButton'
import * as subscribeModule from '../push/subscribe'

vi.mock('../push/subscribe')
vi.mock('sonner', () => ({ toast: { error: vi.fn() } }))

beforeEach(() => {
  vi.clearAllMocks()
  // jsdom does not implement the Notification API at all, so it must be
  // stubbed wholesale rather than patched via Object.defineProperty.
  vi.stubGlobal('Notification', { permission: 'default', requestPermission: vi.fn() })
})

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('NotifyMeButton', () => {
  it('renders nothing when push is unsupported', async () => {
    vi.mocked(subscribeModule.isPushSupported).mockReturnValue(false)

    const { container } = render(<NotifyMeButton />)

    await waitFor(() => expect(container).not.toHaveTextContent('checking'))
    expect(container).toBeEmptyDOMElement()
  })

  it('shows the opt-in button when supported and not yet subscribed', async () => {
    vi.mocked(subscribeModule.isPushSupported).mockReturnValue(true)
    vi.mocked(subscribeModule.getExistingSubscription).mockResolvedValue(null)

    render(<NotifyMeButton />)

    expect(await screen.findByRole('button', { name: /get notified/i })).toBeInTheDocument()
  })

  it('shows "Notifications on" when already subscribed', async () => {
    vi.mocked(subscribeModule.isPushSupported).mockReturnValue(true)
    vi.mocked(subscribeModule.getExistingSubscription).mockResolvedValue({} as unknown as PushSubscription)

    render(<NotifyMeButton />)

    expect(await screen.findByText(/notifications on/i)).toBeInTheDocument()
  })

  it('subscribes on click and switches to the subscribed state', async () => {
    vi.mocked(subscribeModule.isPushSupported).mockReturnValue(true)
    vi.mocked(subscribeModule.getExistingSubscription).mockResolvedValue(null)
    vi.mocked(subscribeModule.subscribeToPush).mockResolvedValue({} as unknown as PushSubscription)

    render(<NotifyMeButton />)
    const button = await screen.findByRole('button', { name: /get notified/i })
    await userEvent.click(button)

    expect(await screen.findByText(/notifications on/i)).toBeInTheDocument()
  })

  it('shows the blocked hint when permission was previously denied', async () => {
    vi.mocked(subscribeModule.isPushSupported).mockReturnValue(true)
    vi.stubGlobal('Notification', { permission: 'denied', requestPermission: vi.fn() })

    render(<NotifyMeButton />)

    expect(await screen.findByText(/blocked/i)).toBeInTheDocument()
  })
})
