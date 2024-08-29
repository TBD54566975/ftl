import { expect, ftlTest } from './ftl-test'

ftlTest('shows infrastructure', async ({ page }) => {
  const infrastructureNavItem = page.getByRole('link', { name: 'Infrastructure' })
  await infrastructureNavItem.click()
  await expect(page).toHaveURL(/\/infrastructure$/)

    const controllersTab = await page.getByRole('button', { name: 'Controllers' });
    await expect(controllersTab).toBeVisible();

    const runnersTab = await page.getByRole('button', { name: 'Runners' });
    await expect(runnersTab).toBeVisible();

    const deploymentsTab = await page.getByRole('button', { name: 'Deployments' });
    await expect(deploymentsTab).toBeVisible();

    const routesTab = await page.getByRole('button', { name: 'Routes' });
    await expect(routesTab).toBeVisible();
})
