import { expect, test } from '@playwright/test'

test('shows active modules', async ({ page }) => {
  await page.goto('http://localhost:8892')
  const modulesNavItem = page.getByRole('link', { name: 'Modules' })
  await modulesNavItem.click()
  await expect(page).toHaveURL(/\/modules$/)

  await expect(page.getByText('dpl-time')).toBeVisible()
  await expect(page.getByText('dpl-echo')).toBeVisible()
})
