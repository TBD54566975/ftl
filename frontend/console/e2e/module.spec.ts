import { expect, ftlTest } from './ftl-test'

ftlTest('shows verbs for deployment', async ({ page }) => {
  const modulesNavItem = page.getByRole('link', { name: 'Modules' })
  await modulesNavItem.click()

  const moduleEchoRow = page.locator('div.cursor-pointer').getByText('echo')
  const moduleEcho = moduleEchoRow.locator('svg').nth(1)
  await moduleEcho.click()

  await expect(page).toHaveURL(/\/modules\/echo/)

  await expect(page.getByText('module echo {')).toBeVisible()
})
