import { test, expect, setupApiMocks, setRoles, dragCardToColumn } from './fixtures'
import { ALL_ORDERS } from './mocks/data'
import type { Order } from '../src/types'

test.describe('Kanban board — layout', () => {
  test.beforeEach(async ({ page }) => {
    await setRoles(page, ['owner'])
    await setupApiMocks(page)
  })

  test('shows Order Board heading', async ({ page }) => {
    await page.goto('/admin/orders')

    await expect(page.getByRole('heading', { name: 'Order Board' })).toBeVisible()
  })

  test('shows all four kanban columns', async ({ page }) => {
    await page.goto('/admin/orders')

    await expect(page.getByTestId('kanban-column')).toHaveCount(4)
    await expect(page.getByText('New Orders')).toBeVisible()
    await expect(page.getByText('Accepted')).toBeVisible()
    await expect(page.getByText('Ongoing')).toBeVisible()
    await expect(page.getByText('Finished')).toBeVisible()
  })

  test('places orders in the correct columns', async ({ page }) => {
    await page.goto('/admin/orders')

    const newOrdersCol = page.getByTestId('kanban-column').filter({ hasText: 'New Orders' })
    const acceptedCol = page.getByTestId('kanban-column').filter({ hasText: 'Accepted' })
    const ongoingCol = page.getByTestId('kanban-column').filter({ hasText: 'Ongoing' })
    const finishedCol = page.getByTestId('kanban-column').filter({ hasText: 'Finished' })

    await expect(newOrdersCol.getByText('Alice')).toBeVisible()
    await expect(newOrdersCol.getByText('Bob')).toBeVisible()
    await expect(acceptedCol.getByText('Carol')).toBeVisible()
    await expect(ongoingCol.getByText('Dave')).toBeVisible()
    await expect(finishedCol.getByText('Eve')).toBeVisible()
  })

  test('shows correct order counts per column', async ({ page }) => {
    await page.goto('/admin/orders')

    const newOrdersCol = page.getByTestId('kanban-column').filter({ hasText: 'New Orders' })
    const acceptedCol = page.getByTestId('kanban-column').filter({ hasText: 'Accepted' })

    await expect(newOrdersCol.getByTestId('column-count')).toHaveText('2')
    await expect(acceptedCol.getByTestId('column-count')).toHaveText('1')
  })

  test('shows order items inside cards', async ({ page }) => {
    await page.goto('/admin/orders')

    const ongoingCol = page.getByTestId('kanban-column').filter({ hasText: 'Ongoing' })
    await expect(ongoingCol.getByText(/Sourdough Loaf/)).toBeVisible()
    await expect(ongoingCol.getByText(/Croissant/)).toBeVisible()
  })

  test('shows "No orders" text in empty column', async ({ page }) => {
    const ordersWithoutFinished = ALL_ORDERS.filter((o) => o.status !== 'finished')
    // Override the orders mock (registered after beforeEach setupApiMocks → takes precedence)
    await page.route('http://localhost:8080/api/v1/orders', (route) => {
      route.fulfill({ json: ordersWithoutFinished })
    })
    await page.goto('/admin/orders')

    const finishedCol = page.getByTestId('kanban-column').filter({ hasText: 'Finished' })
    await expect(finishedCol.getByText('No orders')).toBeVisible()
  })

  test('shows drag instruction text', async ({ page }) => {
    await page.goto('/admin/orders')

    await expect(
      page.getByText('Drag orders left or right to change their status.'),
    ).toBeVisible()
  })
})

test.describe('Kanban board — drag and drop', () => {
  // beforeEach sets roles; each test calls setupApiMocks then adds its own overrides
  test.beforeEach(async ({ page }) => {
    await setRoles(page, ['owner'])
  })

  test('can drag a "New Orders" card to "Accepted"', async ({ page }) => {
    // 1. Set up base mocks first
    await setupApiMocks(page)

    // 2. Override accept route AFTER setupApiMocks (LIFO: this one wins)
    let acceptCalled = false
    await page.route('http://localhost:8080/api/v1/orders/order-new-1/accept', (route) => {
      acceptCalled = true
      const updated: Order = {
        ...ALL_ORDERS.find((o) => o.id === 'order-new-1')!,
        status: 'accepted',
      }
      route.fulfill({ json: updated })
    })

    await page.goto('/admin/orders')

    const aliceCard = page.getByTestId('order-card').filter({ hasText: 'Alice' })
    const acceptedCol = page.getByTestId('kanban-column').filter({ hasText: 'Accepted' })

    await aliceCard.waitFor({ state: 'visible' })
    await acceptedCol.waitFor({ state: 'visible' })

    await dragCardToColumn(page, aliceCard, acceptedCol)

    expect(acceptCalled).toBe(true)
  })

  test('can drag an "Accepted" card to "Ongoing"', async ({ page }) => {
    await setupApiMocks(page)

    let startCalled = false
    await page.route('http://localhost:8080/api/v1/orders/order-accepted-1/start', (route) => {
      startCalled = true
      const updated: Order = {
        ...ALL_ORDERS.find((o) => o.id === 'order-accepted-1')!,
        status: 'ongoing',
      }
      route.fulfill({ json: updated })
    })

    await page.goto('/admin/orders')

    const carolCard = page.getByTestId('order-card').filter({ hasText: 'Carol' })
    const ongoingCol = page.getByTestId('kanban-column').filter({ hasText: 'Ongoing' })

    await carolCard.waitFor({ state: 'visible' })
    await ongoingCol.waitFor({ state: 'visible' })

    await dragCardToColumn(page, carolCard, ongoingCol)

    expect(startCalled).toBe(true)
  })

  test('can drag an "Ongoing" card to "Finished"', async ({ page }) => {
    await setupApiMocks(page)

    let finishCalled = false
    await page.route('http://localhost:8080/api/v1/orders/order-ongoing-1/finish', (route) => {
      finishCalled = true
      const updated: Order = {
        ...ALL_ORDERS.find((o) => o.id === 'order-ongoing-1')!,
        status: 'finished',
      }
      route.fulfill({ json: updated })
    })

    await page.goto('/admin/orders')

    const daveCard = page.getByTestId('order-card').filter({ hasText: 'Dave' })
    const finishedCol = page.getByTestId('kanban-column').filter({ hasText: 'Finished' })

    await daveCard.waitFor({ state: 'visible' })
    await finishedCol.waitFor({ state: 'visible' })

    await dragCardToColumn(page, daveCard, finishedCol)

    expect(finishCalled).toBe(true)
  })

  test('cannot drag a card more than one column forward (skipping)', async ({ page }) => {
    await setupApiMocks(page)

    let startCalled = false
    // Override the start route to detect if it's mistakenly called
    await page.route('http://localhost:8080/api/v1/orders/order-new-1/start', (route) => {
      startCalled = true
      route.fulfill({ json: ALL_ORDERS.find((o) => o.id === 'order-new-1') })
    })
    let acceptCalled = false
    await page.route('http://localhost:8080/api/v1/orders/order-new-1/accept', (route) => {
      acceptCalled = true
      route.fulfill({ json: ALL_ORDERS.find((o) => o.id === 'order-new-1') })
    })

    await page.goto('/admin/orders')

    const aliceCard = page.getByTestId('order-card').filter({ hasText: 'Alice' })
    // Try to drag to "Ongoing" (skipping "Accepted") — app should ignore this
    const ongoingCol = page.getByTestId('kanban-column').filter({ hasText: 'Ongoing' })

    await aliceCard.waitFor({ state: 'visible' })
    await dragCardToColumn(page, aliceCard, ongoingCol)

    // Neither accept nor start should be called (only one step forward allowed)
    expect(acceptCalled).toBe(false)
    expect(startCalled).toBe(false)
  })

  test('can drag an "Accepted" card back to "New Orders"', async ({ page }) => {
    await setupApiMocks(page)

    let unacceptCalled = false
    await page.route('http://localhost:8080/api/v1/orders/order-accepted-1/unaccept', (route) => {
      unacceptCalled = true
      const updated: Order = {
        ...ALL_ORDERS.find((o) => o.id === 'order-accepted-1')!,
        status: 'created',
      }
      route.fulfill({ json: updated })
    })

    await page.goto('/admin/orders')

    const carolCard = page.getByTestId('order-card').filter({ hasText: 'Carol' })
    const newOrdersCol = page.getByTestId('kanban-column').filter({ hasText: 'New Orders' })

    await carolCard.waitFor({ state: 'visible' })
    await dragCardToColumn(page, carolCard, newOrdersCol)

    expect(unacceptCalled).toBe(true)
  })

  test('can drag an "Ongoing" card back to "Accepted"', async ({ page }) => {
    await setupApiMocks(page)

    let stopCalled = false
    await page.route('http://localhost:8080/api/v1/orders/order-ongoing-1/stop', (route) => {
      stopCalled = true
      const updated: Order = {
        ...ALL_ORDERS.find((o) => o.id === 'order-ongoing-1')!,
        status: 'accepted',
      }
      route.fulfill({ json: updated })
    })

    await page.goto('/admin/orders')

    const daveCard = page.getByTestId('order-card').filter({ hasText: 'Dave' })
    const acceptedCol = page.getByTestId('kanban-column').filter({ hasText: 'Accepted' })

    await daveCard.waitFor({ state: 'visible' })
    await dragCardToColumn(page, daveCard, acceptedCol)

    expect(stopCalled).toBe(true)
  })

  test('can drag a "Finished" card back to "Ongoing"', async ({ page }) => {
    await setupApiMocks(page)

    let unfinishCalled = false
    await page.route('http://localhost:8080/api/v1/orders/order-finished-1/unfinish', (route) => {
      unfinishCalled = true
      const updated: Order = {
        ...ALL_ORDERS.find((o) => o.id === 'order-finished-1')!,
        status: 'ongoing',
      }
      route.fulfill({ json: updated })
    })

    await page.goto('/admin/orders')

    const eveCard = page.getByTestId('order-card').filter({ hasText: 'Eve' })
    const ongoingCol = page.getByTestId('kanban-column').filter({ hasText: 'Ongoing' })

    await eveCard.waitFor({ state: 'visible' })
    await dragCardToColumn(page, eveCard, ongoingCol)

    expect(unfinishCalled).toBe(true)
  })

  test('cannot drag a card more than one column backward (skipping)', async ({ page }) => {
    await setupApiMocks(page)

    let mutationCalled = false
    await page.route('http://localhost:8080/api/v1/orders/order-ongoing-1/**', (route) => {
      if (route.request().method() !== 'GET') {
        mutationCalled = true
        route.fulfill({ json: ALL_ORDERS.find((o) => o.id === 'order-ongoing-1') })
      } else {
        route.continue()
      }
    })

    await page.goto('/admin/orders')

    const daveCard = page.getByTestId('order-card').filter({ hasText: 'Dave' })
    // Try to drag Dave (Ongoing) back two steps to New Orders — should be ignored
    const newOrdersCol = page.getByTestId('kanban-column').filter({ hasText: 'New Orders' })

    await daveCard.waitFor({ state: 'visible' })
    await dragCardToColumn(page, daveCard, newOrdersCol)

    expect(mutationCalled).toBe(false)
  })
})

test.describe('Kanban board — access control', () => {
  test('non-owner is redirected away from kanban', async ({ page }) => {
    await setRoles(page, [])
    await setupApiMocks(page)
    await page.goto('/admin/orders')

    // The component calls navigate('/') on render when not owner
    await page.waitForURL('/')
    await expect(page).toHaveURL('/')
  })
})

test.describe('Kanban board — error states', () => {
  test.beforeEach(async ({ page }) => {
    await setRoles(page, ['owner'])
  })

  test('shows error state when API fails', async ({ page }) => {
    await page.route('http://localhost:8080/api/v1/orders', (route) => {
      route.fulfill({ status: 500, json: { error: 'server error' } })
    })
    await page.goto('/admin/orders')

    // retry: false in test mode so error state appears immediately
    await expect(page.getByText('Could not load orders. Try again.')).toBeVisible()
  })

  test('shows loading skeleton initially', async ({ page }) => {
    let resolveOrders!: () => void
    const ordersReady = new Promise<void>((res) => {
      resolveOrders = res
    })

    await page.route('http://localhost:8080/api/v1/orders', async (route) => {
      await ordersReady
      route.fulfill({ json: ALL_ORDERS })
    })

    await page.goto('/admin/orders')
    // The loading state should show the animate-pulse skeleton
    await expect(page.getByTestId('loading-skeleton')).toBeVisible()

    resolveOrders()
  })
})
