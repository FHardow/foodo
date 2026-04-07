import { test, expect, setupApiMocks, setRoles } from './fixtures'
import { BASKET_ORDER, BASKET_ORDER_WITH_ITEMS } from './mocks/data'
import type { Page } from '@playwright/test'

// Zustand persist uses the `name` option as the localStorage key
const BASKET_LS_KEY = 'basketOrderId'

function setBasketInStorage(page: Page, orderId: string) {
  return page.addInitScript(
    ({ key, id }: { key: string; id: string }) => {
      localStorage.setItem(key, JSON.stringify({ state: { basketOrderId: id }, version: 0 }))
    },
    { key: BASKET_LS_KEY, id: orderId },
  )
}

test.describe('Basket — empty state', () => {
  test.beforeEach(async ({ page }) => {
    await setRoles(page, [])
    await setupApiMocks(page, { basketOrder: null })
  })

  test('shows empty basket message', async ({ page }) => {
    await page.goto('/basket')

    await expect(page.getByText('Your basket is empty')).toBeVisible()
    await expect(page.getByRole('link', { name: 'Browse products' })).toBeVisible()
  })

  test('browse products link goes to store', async ({ page }) => {
    await page.goto('/basket')

    await page.getByRole('link', { name: 'Browse products' }).click()
    await expect(page).toHaveURL('/')
  })
})

test.describe('Basket — with items', () => {
  test.beforeEach(async ({ page }) => {
    await setRoles(page, [])
    await setBasketInStorage(page, BASKET_ORDER_WITH_ITEMS.id)
    await setupApiMocks(page, { basketOrder: BASKET_ORDER_WITH_ITEMS })
  })

  test('shows basket heading', async ({ page }) => {
    await page.goto('/basket')

    await expect(page.getByRole('heading', { name: 'Your Basket' })).toBeVisible()
  })

  test('shows items in basket', async ({ page }) => {
    await page.goto('/basket')

    await expect(page.getByText('Sourdough Loaf')).toBeVisible()
    await expect(page.getByText('Croissant')).toBeVisible()
  })

  test('shows quantity controls for each item', async ({ page }) => {
    await page.goto('/basket')

    // Each item has +/- buttons and a Remove button
    await expect(page.getByRole('button', { name: '−' })).toHaveCount(2)
    await expect(page.getByRole('button', { name: '+' })).toHaveCount(2)
    await expect(page.getByRole('button', { name: 'Remove' })).toHaveCount(2)
  })

  test('shows Place Order button', async ({ page }) => {
    await page.goto('/basket')

    await expect(page.getByRole('button', { name: 'Place Order' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Place Order' })).toBeEnabled()
  })

  test('Place Order navigates to order status page', async ({ page }) => {
    // Mock the placed order status endpoint (registered before setupApiMocks so lower priority,
    // but we also override it after to be safe)
    await page.goto('/basket')
    await expect(page.getByRole('button', { name: 'Place Order' })).toBeVisible()

    // The confirm route is already mocked by setupApiMocks to return placed-order-id
    // Also mock the placed order GET
    await page.route('http://localhost:8080/api/v1/orders/placed-order-id', (route) => {
      route.fulfill({
        json: { ...BASKET_ORDER_WITH_ITEMS, id: 'placed-order-id', status: 'created' },
      })
    })

    await page.getByRole('button', { name: 'Place Order' }).click()
    await expect(page).toHaveURL(/\/orders\/placed-order-id/)
  })
})

test.describe('Basket — empty order (no items)', () => {
  test('Place Order button is disabled when basket has no items', async ({ page }) => {
    await setRoles(page, [])
    await setBasketInStorage(page, BASKET_ORDER.id)
    await setupApiMocks(page, { basketOrder: BASKET_ORDER })

    await page.goto('/basket')

    await expect(page.getByRole('button', { name: 'Place Order' })).toBeDisabled()
  })
})
