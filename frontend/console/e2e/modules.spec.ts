import { expect, test } from '@playwright/test'

test('shows active modules', async ({ page }) => {
  await page.goto('/')
  const modulesNavItem = page.getByRole('link', { name: 'Modules' })
  await modulesNavItem.click()
  await expect(page).toHaveURL(/\/modules$/)

  await page.waitForSelector('[data-module-row]')
  const moduleNames = await page.$$eval('[data-module-row]', (elements) => elements.map((el) => el.getAttribute('data-module-row')))

  const expectedModuleNames = ['cron', 'time', 'pubsub', 'http', 'echo']
  expect(moduleNames).toEqual(expect.arrayContaining(expectedModuleNames))
})

test('tapping on a module navigates to the module page', async ({ page }) => {
  await page.goto('/modules')
  await page.locator(`[data-module-row="echo"]`).click()
  await expect(page).toHaveURL(/\/modules\/echo$/)
})
