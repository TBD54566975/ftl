import { expect, test } from '@playwright/test'
import { navigateToDecl } from './helpers'

test('shows cron verb form', async ({ page }) => {
  await navigateToDecl(page, 'cron', 'thirtySeconds')

  await expect(page.getByText('CRON', { exact: true })).toBeVisible()
  await expect(page.locator('input#request-path')).toHaveValue('cron.thirtySeconds')

  await expect(page.getByText('Body', { exact: true })).toBeVisible()
  await expect(page.getByText('Schema', { exact: true })).toBeVisible()
})

test('send cron request', async ({ page }) => {
  await navigateToDecl(page, 'cron', 'thirtySeconds')

  await page.getByRole('button', { name: 'Send' }).click()

  const responseEditor = page.locator('#response-editor .cm-content[role="textbox"]')
  await expect(responseEditor).toBeVisible()

  const responseText = await responseEditor.textContent()
  const responseJson = JSON.parse(responseText?.trim() || '{}')

  expect(responseJson).toEqual({})
})

test('submit cron form using ⌥+⏎ shortcut', async ({ page }) => {
  await navigateToDecl(page, 'cron', 'thirtySeconds')

  await page.locator('input#request-path').focus()

  await page.keyboard.press('Alt+Enter')

  const responseEditor = page.locator('#response-editor .cm-content[role="textbox"]')
  await expect(responseEditor).toBeVisible()

  const responseText = await responseEditor.textContent()
  const responseJson = JSON.parse(responseText?.trim() || '{}')

  expect(responseJson).toEqual({})
})

test('submit cron form using ⌥+⏎ shortcut without focusing first', async ({ page }) => {
  await navigateToDecl(page, 'cron', 'thirtySeconds')

  await page.keyboard.press('Alt+Enter')

  const responseEditor = page.locator('#response-editor .cm-content[role="textbox"]')
  await expect(responseEditor).toBeVisible()

  const responseText = await responseEditor.textContent()
  const responseJson = JSON.parse(responseText?.trim() || '{}')

  expect(responseJson).toEqual({})
})
