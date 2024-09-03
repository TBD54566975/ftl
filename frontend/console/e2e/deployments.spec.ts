import { expect, ftlTest } from './ftl-test'

ftlTest('shows active deployments', async ({ page }) => {
  const deploymentsNavItem = page.getByRole('link', { name: 'Deployments' })
  await deploymentsNavItem.click()
  await expect(page).toHaveURL(/\/deployments$/)

  await expect(page.getByText('dpl-time')).toBeVisible()
  await expect(page.getByText('dpl-echo')).toBeVisible()
})
