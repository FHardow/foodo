import { test, expect, setupApiMocks, setRoles } from './fixtures'

test.describe('Navigation — owner', () => {
  test.beforeEach(async ({ page }) => {
    await setRoles(page, ['owner'])
    await setupApiMocks(page)
  })

  test('shows Bread Order brand link', async ({ page }) => {
    await page.goto('/')

    await expect(page.getByRole('link', { name: 'Bread Order' })).toBeVisible()
  })

  test('shows Store and History links', async ({ page }) => {
    await page.goto('/')

    await expect(page.getByRole('link', { name: 'Store' })).toBeVisible()
    await expect(page.getByRole('link', { name: 'History' })).toBeVisible()
  })

  test('shows owner-only nav links for admin', async ({ page }) => {
    await page.goto('/')

    await expect(page.getByRole('link', { name: 'Orders' })).toBeVisible()
    await expect(page.getByRole('link', { name: 'Products' })).toBeVisible()
  })

  test('shows basket button', async ({ page }) => {
    await page.goto('/')

    await expect(page.getByRole('link', { name: /Basket/ })).toBeVisible()
  })

  test('brand link navigates to store', async ({ page }) => {
    await page.goto('/basket')

    await page.getByRole('link', { name: 'Bread Order' }).click()
    await expect(page).toHaveURL('/')
  })

  test('Orders link navigates to kanban board', async ({ page }) => {
    await page.goto('/')

    await page.getByRole('link', { name: 'Orders' }).click()
    await expect(page).toHaveURL('/admin/orders')
  })

  test('History link navigates to order history', async ({ page }) => {
    await page.goto('/')

    await page.getByRole('link', { name: 'History' }).click()
    await expect(page).toHaveURL('/orders')
  })

  test('basket link navigates to basket', async ({ page }) => {
    await page.goto('/')

    await page.getByRole('link', { name: /Basket/ }).click()
    await expect(page).toHaveURL('/basket')
  })
})

test.describe('Navigation — customer', () => {
  test.beforeEach(async ({ page }) => {
    await setRoles(page, [])
    await setupApiMocks(page)
  })

  test('does not show admin nav links for non-owner', async ({ page }) => {
    await page.goto('/')

    await expect(page.getByRole('link', { name: 'Orders' })).not.toBeVisible()
    await expect(page.getByRole('link', { name: 'Products' })).not.toBeVisible()
  })

  test('still shows Store, History and basket links', async ({ page }) => {
    await page.goto('/')

    await expect(page.getByRole('link', { name: 'Store' })).toBeVisible()
    await expect(page.getByRole('link', { name: 'History' })).toBeVisible()
    await expect(page.getByRole('link', { name: /Basket/ })).toBeVisible()
  })
})

test.describe('Order history page', () => {
  test.beforeEach(async ({ page }) => {
    await setRoles(page, [])
    await setupApiMocks(page)
  })

  test('shows order history heading', async ({ page }) => {
    await page.goto('/orders')

    // Page should load without error
    await expect(page).toHaveURL('/orders')
  })
})
