import { expect, ftlTest } from './ftl-test'

ftlTest('shows verbs for deployment', async ({ page }) => {
  const deploymentsNavItem = page.getByRole('link', { name: 'Deployments' })
  await deploymentsNavItem.click()

  const deploymentEcho = page.getByText('dpl-echo')
  await deploymentEcho.click()

  await expect(page).toHaveURL(/\/deployments\/dpl-echo-.*/)

  await expect(page.getByText('echo', { exact: true })).toBeVisible()
  await expect(page.getByText('exported', { exact: true })).toBeVisible()
})
