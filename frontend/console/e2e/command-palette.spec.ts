import { expect, test } from '@playwright/test'
import { pressShortcut } from './helpers'

test('shows command palette results', async ({ page }) => {
  await page.goto('/')

  await page.click('#command-palette-search')
  await page.fill('#command-palette-search-input', 'echo')

  // Command palette should be showing the echo parts
  await expect(page.locator('[role="listbox"]').getByText('echo.EchoRequest', { exact: true })).toBeVisible()
  await expect(page.locator('[role="listbox"]').getByText('echo.EchoResponse', { exact: true })).toBeVisible()
  await expect(page.locator('[role="listbox"]').getByText('echo.echo', { exact: true })).toBeVisible()
})

test('opens command palette with keyboard shortcut', async ({ page }) => {
  await page.goto('/')

  await pressShortcut(page, 'k')

  await expect(page.locator('#command-palette-search-input')).toBeVisible()
  await page.fill('#command-palette-search-input', 'echo')

  // Command palette should be showing the echo parts
  await expect(page.locator('[role="listbox"]').getByText('echo.EchoRequest', { exact: true })).toBeVisible()
  await expect(page.locator('[role="listbox"]').getByText('echo.EchoResponse', { exact: true })).toBeVisible()
  await expect(page.locator('[role="listbox"]').getByText('echo.echo', { exact: true })).toBeVisible()
})
