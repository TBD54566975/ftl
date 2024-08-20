import { expect, ftlTest } from './ftl-test'

ftlTest.beforeEach(async ({ page }) => {
  const deploymentsNavItem = page.getByRole('link', { name: 'Deployments' })
  await deploymentsNavItem.click()

  const deploymentEcho = page.getByText('dpl-echo')
  await deploymentEcho.click()

  const verbEcho = page.getByText('echo', { exact: true })
  await verbEcho.click()

  await expect(page).toHaveURL(/\/deployments\/dpl-echo-[^/]+\/verbs\/echo/)
})

ftlTest('shows verb form', async ({ page }) => {
  await expect(page.getByText('CALL', { exact: true })).toBeVisible()
  await expect(page.locator('input#request-path')).toHaveValue('echo.echo')

  await expect(page.getByText('Body', { exact: true })).toBeVisible()
  await expect(page.getByText('Verb Schema', { exact: true })).toBeVisible()
  await expect(page.getByText('JSONSchema', { exact: true })).toBeVisible()
})
