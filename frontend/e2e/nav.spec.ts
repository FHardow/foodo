import { test, expect, setupApiMocks, setRoles } from './fixtures'

test.describe('Navigation — owner', () => {
  test.beforeEach(async ({ page }) => {
    await setRoles(page, ['owner'])
    await setupApiMocks(page)
  })

  test('shows Foodo brand link', async ({ page }) => {
    await page.goto('/')

    await expect(page.getByRole('link', { name: 'Foodo' })).toBeVisible()
  })

  test('shows Store and History links', async ({ page, isMobile }) => {
    test.skip(isMobile, 'Desktop-only row — on mobile these live inside the burger menu')
    await page.goto('/')

    await expect(page.getByRole('link', { name: 'Store' })).toBeVisible()
    await expect(page.getByRole('link', { name: 'History' })).toBeVisible()
  })

  test('shows owner-only nav links for admin', async ({ page, isMobile }) => {
    test.skip(isMobile, 'Desktop-only row — on mobile these live inside the burger menu')
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

    await page.getByRole('link', { name: 'Foodo' }).click()
    await expect(page).toHaveURL('/')
  })

  test('Orders link navigates to kanban board', async ({ page, isMobile }) => {
    test.skip(isMobile, 'Desktop-only row — on mobile these live inside the burger menu')
    await page.goto('/')

    await page.getByRole('link', { name: 'Orders' }).click()
    await expect(page).toHaveURL('/admin/orders')
  })

  test('History link navigates to order history', async ({ page, isMobile }) => {
    test.skip(isMobile, 'Desktop-only row — on mobile these live inside the burger menu')
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

  test('still shows Store, History and basket links', async ({ page, isMobile }) => {
    test.skip(isMobile, 'Desktop-only row — on mobile these live inside the burger menu')
    await page.goto('/')

    await expect(page.getByRole('link', { name: 'Store' })).toBeVisible()
    await expect(page.getByRole('link', { name: 'History' })).toBeVisible()
    await expect(page.getByRole('link', { name: /Basket/ })).toBeVisible()
  })
})

test.describe('Navigation — mobile burger menu', () => {
  test.beforeEach(async ({ page }) => {
    await setRoles(page, ['owner'])
    await setupApiMocks(page)
  })

  test('shows a menu button on mobile; desktop links stay hidden', async ({ page, isMobile }) => {
    test.skip(!isMobile, 'Burger menu only renders below the sm breakpoint')
    await page.goto('/')

    await expect(page.getByRole('button', { name: 'Open menu' })).toBeVisible()
    await expect(page.getByRole('navigation').getByRole('link', { name: 'Store' })).not.toBeVisible()
  })

  test('opens the menu and shows all nav links, including owner-only ones', async ({ page, isMobile }) => {
    test.skip(!isMobile, 'Burger menu only renders below the sm breakpoint')
    await page.goto('/')

    await page.getByRole('button', { name: 'Open menu' }).click()
    const menu = page.getByRole('dialog')
    await expect(menu.getByRole('link', { name: 'Store' })).toBeVisible()
    await expect(menu.getByRole('link', { name: 'History' })).toBeVisible()
    await expect(menu.getByRole('link', { name: 'Orders' })).toBeVisible()
    await expect(menu.getByRole('link', { name: 'Products' })).toBeVisible()
  })

  test('clicking a link in the menu navigates and closes it', async ({ page, isMobile }) => {
    test.skip(!isMobile, 'Burger menu only renders below the sm breakpoint')
    await page.goto('/')

    await page.getByRole('button', { name: 'Open menu' }).click()
    await page.getByRole('dialog').getByRole('link', { name: 'History' }).click()

    await expect(page).toHaveURL('/orders')
    await expect(page.getByRole('dialog')).not.toBeVisible()
  })

  test('non-owner does not see owner-only links in the menu', async ({ page, isMobile }) => {
    test.skip(!isMobile, 'Burger menu only renders below the sm breakpoint')
    await setRoles(page, [])
    await setupApiMocks(page)
    await page.goto('/')

    await page.getByRole('button', { name: 'Open menu' }).click()
    const menu = page.getByRole('dialog')
    await expect(menu.getByRole('link', { name: 'Orders' })).toHaveCount(0)
    await expect(menu.getByRole('link', { name: 'Products' })).toHaveCount(0)
  })
})

test.describe('Order history page', () => {
  test.beforeEach(async ({ page }) => {
    await setRoles(page, [])
    await setupApiMocks(page)
  })

  test('shows order history heading', async ({ page }) => {
    await page.goto('/orders')

    await expect(page).toHaveURL('/orders')
  })
})
