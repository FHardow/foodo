import { test as base, expect, type Page } from '@playwright/test'
import type { Order, Product } from '../../src/types'
import {
  PRODUCTS,
  BASKET_ORDER,
  ALL_ORDERS,
} from '../mocks/data'

export { expect }

const API = 'http://localhost:8080'

/**
 * Set up API route mocks. Call before page.goto().
 *
 * Playwright uses LIFO ordering: routes registered LAST take precedence.
 * We register routes from most-general to most-specific so the specific
 * ones win. Callers who want to override a route should call their
 * page.route() AFTER this function.
 */
export async function setupApiMocks(
  page: Page,
  overrides: {
    products?: Product[]
    allOrders?: Order[]
    basketOrder?: Order | null
  } = {},
) {
  const products = overrides.products ?? PRODUCTS
  const allOrders = overrides.allOrders ?? ALL_ORDERS
  const basketOrder = overrides.basketOrder !== undefined ? overrides.basketOrder : null

  // ── Mutation routes (registered first = lowest LIFO priority) ──────────

  await page.route(`${API}/api/v1/orders/*/confirm`, (route) => {
    if (route.request().method() === 'POST') {
      route.fulfill({
        json: { ...(basketOrder ?? BASKET_ORDER), id: 'placed-order-id', status: 'created' },
      })
    } else {
      route.continue()
    }
  })

  await page.route(`${API}/api/v1/orders/*/items/*`, (route) => {
    if (route.request().method() === 'DELETE') {
      route.fulfill({ json: basketOrder ?? BASKET_ORDER })
    } else {
      route.continue()
    }
  })

  await page.route(`${API}/api/v1/orders/*/items`, (route) => {
    if (route.request().method() === 'POST') {
      route.fulfill({ json: basketOrder ?? BASKET_ORDER })
    } else {
      route.continue()
    }
  })

  await page.route(`${API}/api/v1/orders/*/unfinish`, (route) => {
    const orderId = route.request().url().split('/orders/')[1]?.split('/')[0]
    const order = allOrders.find((o) => o.id === orderId)
    route.fulfill({ json: { ...(order ?? {}), status: 'ongoing' } })
  })

  await page.route(`${API}/api/v1/orders/*/stop`, (route) => {
    const orderId = route.request().url().split('/orders/')[1]?.split('/')[0]
    const order = allOrders.find((o) => o.id === orderId)
    route.fulfill({ json: { ...(order ?? {}), status: 'accepted' } })
  })

  await page.route(`${API}/api/v1/orders/*/unaccept`, (route) => {
    const orderId = route.request().url().split('/orders/')[1]?.split('/')[0]
    const order = allOrders.find((o) => o.id === orderId)
    route.fulfill({ json: { ...(order ?? {}), status: 'created' } })
  })

  await page.route(`${API}/api/v1/orders/*/finish`, (route) => {
    const orderId = route.request().url().split('/orders/')[1]?.split('/')[0]
    const order = allOrders.find((o) => o.id === orderId)
    route.fulfill({ json: { ...(order ?? {}), status: 'finished' } })
  })

  await page.route(`${API}/api/v1/orders/*/start`, (route) => {
    const orderId = route.request().url().split('/orders/')[1]?.split('/')[0]
    const order = allOrders.find((o) => o.id === orderId)
    route.fulfill({ json: { ...(order ?? {}), status: 'ongoing' } })
  })

  await page.route(`${API}/api/v1/orders/*/accept`, (route) => {
    const orderId = route.request().url().split('/orders/')[1]?.split('/')[0]
    const order = allOrders.find((o) => o.id === orderId)
    route.fulfill({ json: { ...(order ?? {}), status: 'accepted' } })
  })

  // ── Wildcard individual order GET (registered before specific routes) ──

  await page.route(`${API}/api/v1/orders/*`, (route) => {
    if (route.request().method() !== 'GET') {
      route.continue()
      return
    }
    const orderId = route.request().url().split('/orders/')[1]?.split('/')[0]
    // Check both allOrders and the basket order
    const order =
      allOrders.find((o) => o.id === orderId) ??
      (basketOrder?.id === orderId ? basketOrder : null)
    if (order) {
      route.fulfill({ json: order })
    } else {
      route.fulfill({ status: 404, json: { error: 'not found' } })
    }
  })

  // ── Specific basket order (registered after wildcard → higher priority) ─

  if (basketOrder) {
    await page.route(`${API}/api/v1/orders/${basketOrder.id}`, (route) => {
      route.fulfill({ json: basketOrder })
    })
  }

  // ── Orders list & create (GET / POST) ─────────────────────────────────

  await page.route(`${API}/api/v1/orders`, (route) => {
    const method = route.request().method()
    const url = route.request().url()

    if (method === 'POST') {
      route.fulfill({ json: BASKET_ORDER })
      return
    }

    if (url.includes('user_id=')) {
      route.fulfill({ json: allOrders.filter((o) => o.status !== 'pending') })
    } else {
      route.fulfill({ json: allOrders })
    }
  })

  // ── Products ──────────────────────────────────────────────────────────

  await page.route(`${API}/api/v1/products`, (route) => {
    route.fulfill({ json: products })
  })
}

/**
 * Set window.__e2eRoles before the page loads.
 */
export async function setRoles(page: Page, roles: string[]) {
  await page.addInitScript((r) => {
    window.__e2eRoles = r as string[]
  }, roles)
}

/**
 * Custom test fixtures.
 */
export const test = base.extend<{
  ownerPage: Page
  customerPage: Page
}>({
  ownerPage: async ({ page }, use) => {
    await setRoles(page, ['owner'])
    await use(page)
  },

  customerPage: async ({ page }, use) => {
    await setRoles(page, [])
    await use(page)
  },
})

/**
 * Drag a kanban card to another column using pointer events.
 * @dnd-kit uses PointerSensor with 5px activation distance.
 *
 * On mobile (single-column layout) the target column can be below the fold.
 * We compute absolute page positions for both elements and scroll so that
 * both fit inside the viewport before starting the drag.
 */
export async function dragCardToColumn(
  page: Page,
  cardLocator: ReturnType<Page['locator']>,
  columnLocator: ReturnType<Page['locator']>,
) {
  // Helper: absolute page coords (viewport-relative + scroll offset)
  const absBox = (locator: ReturnType<Page['locator']>) =>
    locator.evaluate((el) => {
      const r = el.getBoundingClientRect()
      return { x: r.x + window.scrollX, y: r.y + window.scrollY, width: r.width, height: r.height }
    })

  const cardAbs = await absBox(cardLocator)
  const colAbs  = await absBox(columnLocator)

  // Scroll so both elements are visible simultaneously
  const viewport    = page.viewportSize()!
  const topEdge     = Math.min(cardAbs.y, colAbs.y)
  const bottomEdge  = Math.max(cardAbs.y + cardAbs.height, colAbs.y + colAbs.height)
  const rangeHeight = bottomEdge - topEdge
  const scrollY = rangeHeight <= viewport.height
    ? Math.max(0, topEdge - (viewport.height - rangeHeight) / 2)
    : Math.max(0, topEdge)
  await page.evaluate((y) => window.scrollTo(0, y), scrollY)
  await page.waitForTimeout(50)

  // Now get viewport-relative bounding boxes (both elements are in view)
  const cardBox = await cardLocator.boundingBox()
  const colBox  = await columnLocator.boundingBox()

  if (!cardBox || !colBox) {
    throw new Error('dragCardToColumn: could not get bounding boxes')
  }

  const startX = cardBox.x + cardBox.width / 2
  const startY = cardBox.y + cardBox.height / 2
  const endX   = colBox.x + colBox.width / 2
  // Target the drop zone area (below the column header)
  const endY   = colBox.y + 80

  // Move to start position
  await page.mouse.move(startX, startY)
  // Press down
  await page.mouse.down()
  // Move slightly to activate the drag sensor (> 5px threshold)
  await page.mouse.move(startX + 8, startY + 8, { steps: 3 })
  // Move to the target column in multiple steps for smooth tracking
  await page.mouse.move(endX, endY, { steps: 30 })
  // Small pause to let dnd-kit register the drop target
  await page.waitForTimeout(100)
  // Release
  await page.mouse.up()
  // Wait for any mutation/refetch to settle
  await page.waitForTimeout(300)
}
