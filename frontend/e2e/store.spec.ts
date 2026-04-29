import { test, expect, setupApiMocks, setRoles } from './fixtures'
import { PRODUCTS } from './mocks/data'

test.describe('Store page', () => {
  test.beforeEach(async ({ page }) => {
    await setRoles(page, ['owner'])
    await setupApiMocks(page)
  })

  test('shows available products in a grid', async ({ page }) => {
    await page.goto('/')

    await expect(page.getByRole('heading', { name: 'Sourdough Loaf' })).toBeVisible()
    await expect(page.getByRole('heading', { name: 'Croissant' })).toBeVisible()
    await expect(page.getByRole('heading', { name: 'Baguette' })).toBeVisible()
  })

  test('does not show unavailable products', async ({ page }) => {
    await page.goto('/')

    await expect(page.getByText('Rye Bread')).not.toBeVisible()
  })

  test('shows product descriptions and units', async ({ page }) => {
    await page.goto('/')

    await expect(page.getByText('Classic tangy sourdough with a crispy crust')).toBeVisible()
    await expect(page.getByText('1 loaf').first()).toBeVisible()
  })

  test('each product card has an Add button', async ({ page }) => {
    await page.goto('/')

    const addButtons = page.getByRole('button', { name: 'Add' })
    await expect(addButtons).toHaveCount(PRODUCTS.filter((p) => p.available).length)
  })

  test('clicking a product card navigates to product detail', async ({ page }) => {
    await page.goto('/')

    await page.getByText('Sourdough Loaf').first().click()
    await expect(page).toHaveURL(/\/products\/prod-1/)
  })
})

test.describe('Store page — no products', () => {
  test('shows empty state when no products available', async ({ page }) => {
    await setRoles(page, ['owner'])
    await setupApiMocks(page, { products: [] })
    await page.goto('/')

    await expect(page.getByText('No products available yet.')).toBeVisible()
  })

  test('shows retry option on API error', async ({ page }) => {
    await setRoles(page, ['owner'])
    // Set up base mocks first, then override products with a 500
    await setupApiMocks(page)
    await page.route('http://localhost:8080/api/v1/products', (route) => {
      route.fulfill({ status: 500, json: { error: 'server error' } })
    })
    await page.goto('/')

    // retry: false in test mode so error state appears immediately
    await expect(page.getByText('Could not load products. Try again.')).toBeVisible()
    await expect(page.getByRole('button', { name: 'Retry' })).toBeVisible()
  })
})

test.describe('Product detail page', () => {
  test.beforeEach(async ({ page }) => {
    await setRoles(page, ['owner'])
    await setupApiMocks(page)
    await page.route('http://localhost:8080/api/v1/products/prod-1', (route) => {
      route.fulfill({ json: PRODUCTS[0] })
    })
  })

  test('shows product name and description', async ({ page }) => {
    await page.goto('/products/prod-1')

    await expect(page.getByText('Sourdough Loaf')).toBeVisible()
    await expect(page.getByText('Classic tangy sourdough with a crispy crust')).toBeVisible()
  })
})
