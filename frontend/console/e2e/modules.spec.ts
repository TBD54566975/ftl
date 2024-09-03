import { expect, ftlTest } from './ftl-test'

ftlTest('shows active modules', async ({ page }) => {
  const modulesNavItem = page.getByRole('link', { name: 'Modules' })
  await modulesNavItem.click()
  await expect(page).toHaveURL(/\/modules$/)

  await expect(page.getByText('dpl-time')).toBeVisible()
  await expect(page.getByText('dpl-echo')).toBeVisible()
})
