import { test, expect, setupApiMocks, setRoles } from './fixtures'
import { PRODUCTS } from './mocks/data'

test.describe('Admin Products — layout', () => {
  test.beforeEach(async ({ page }) => {
    await setRoles(page, ['owner'])
    await setupApiMocks(page)
  })

  test('shows Products heading and existing products', async ({ page }) => {
    await page.goto('/admin/products')

    await expect(page.getByRole('heading', { name: 'Products' })).toBeVisible()
    await expect(page.getByRole('cell', { name: 'Sourdough Loaf' })).toBeVisible()
    await expect(page.getByRole('cell', { name: 'Croissant' })).toBeVisible()
  })

  test('shows an availability switch matching each product state', async ({ page }) => {
    await page.goto('/admin/products')

    await expect(page.getByRole('switch', { name: 'Mark unavailable' })).toHaveCount(
      PRODUCTS.filter((p) => p.available).length,
    )
    await expect(page.getByRole('switch', { name: 'Mark available' })).toHaveCount(
      PRODUCTS.filter((p) => !p.available).length,
    )
  })
})

test.describe('Admin Products — create', () => {
  test.beforeEach(async ({ page }) => {
    await setRoles(page, ['owner'])
    await setupApiMocks(page)
  })

  test('opens an empty dialog when adding a product', async ({ page }) => {
    await page.goto('/admin/products')

    await page.getByRole('button', { name: '+ Add product' }).click()

    const dialog = page.getByRole('dialog')
    await expect(dialog.getByRole('heading', { name: 'Add product' })).toBeVisible()
    await expect(dialog.getByLabel('Name')).toHaveValue('')
    await expect(dialog.getByLabel('Unit')).toHaveValue('')
  })

  test('submitting the form creates a product and closes the dialog', async ({ page }) => {
    let createCalled = false
    await page.route('http://localhost:8080/api/v1/products', (route) => {
      if (route.request().method() === 'POST') {
        createCalled = true
        route.fulfill({
          json: { id: 'prod-new', name: 'Rye Bread', description: '', unit: 'loaf', available: false },
        })
      } else {
        route.fallback()
      }
    })

    await page.goto('/admin/products')
    await page.getByRole('button', { name: '+ Add product' }).click()

    const dialog = page.getByRole('dialog')
    await dialog.getByLabel('Name').fill('Rye Bread')
    await dialog.getByLabel('Unit').fill('loaf')
    await dialog.getByRole('button', { name: 'Create product' }).click()

    await expect(dialog).not.toBeVisible()
    expect(createCalled).toBe(true)
  })
})

test.describe('Admin Products — edit', () => {
  test.beforeEach(async ({ page }) => {
    await setRoles(page, ['owner'])
    await setupApiMocks(page)
  })

  test('opens the dialog pre-filled with the product being edited', async ({ page }) => {
    await page.goto('/admin/products')

    await page
      .getByRole('row', { name: /Sourdough Loaf/ })
      .getByRole('button', { name: 'Edit' })
      .click()

    const dialog = page.getByRole('dialog')
    await expect(dialog.getByRole('heading', { name: 'Edit product' })).toBeVisible()
    await expect(dialog.getByLabel('Name')).toHaveValue('Sourdough Loaf')
    await expect(dialog.getByLabel('Unit')).toHaveValue('1 loaf')
  })
})

test.describe('Admin Products — availability toggle', () => {
  test.beforeEach(async ({ page }) => {
    await setRoles(page, ['owner'])
    await setupApiMocks(page)
  })

  test('toggling the switch calls the availability endpoint', async ({ page }) => {
    let availabilityCalled = false
    await page.route('http://localhost:8080/api/v1/products/prod-1/availability', (route) => {
      availabilityCalled = true
      route.fulfill({ json: { ...PRODUCTS[0], available: false } })
    })

    await page.goto('/admin/products')

    await page.getByRole('switch', { name: 'Mark unavailable' }).first().click()

    expect(availabilityCalled).toBe(true)
  })
})

test.describe('Admin Products — access control', () => {
  test('non-owner is redirected away from admin products', async ({ page }) => {
    await setRoles(page, [])
    await setupApiMocks(page)
    await page.goto('/admin/products')

    await page.waitForURL('/')
    await expect(page).toHaveURL('/')
  })
})
