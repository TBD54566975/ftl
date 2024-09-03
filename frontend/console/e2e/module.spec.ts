import { expect, ftlTest } from './ftl-test'

ftlTest('shows verbs for deployment', async ({ page }) => {
  const modulesNavItem = page.getByRole('link', { name: 'Modules' })
  await modulesNavItem.click()

  const moduleEchoRow = page.getByRole('button', { name: 'echo' })
  const moduleEcho = moduleEchoRow.locator('svg').nth(1)
  await moduleEcho.click()

  await expect(page).toHaveURL(/\/modules\/echo/)

  await expect(page.getByText('Deployment', { exact: true })).toBeVisible()
  await expect(page.getByText('Deployed dpl-echo')).toBeVisible()
})
