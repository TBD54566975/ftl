import { expect, test } from '@playwright/test'

test('shows infrastructure', async ({ page }) => {
  await page.goto('http://localhost:8892')
  const infrastructureNavItem = page.getByRole('link', { name: 'Infrastructure' })
  await infrastructureNavItem.click()
  await expect(page).toHaveURL(/\/infrastructure\/controllers$/)

  const controllersTab = await page.getByRole('button', { name: 'Controllers' })
  await expect(controllersTab).toBeVisible()

  const runnersTab = await page.getByRole('button', { name: 'Runners' })
  await expect(runnersTab).toBeVisible()

  const deploymentsTab = await page.getByRole('button', { name: 'Deployments' })
  await expect(deploymentsTab).toBeVisible()

  const routesTab = await page.getByRole('button', { name: 'Routes' })
  await expect(routesTab).toBeVisible()
})
