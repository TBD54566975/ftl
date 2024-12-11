import { expect, test } from '@playwright/test'
import { navigateToDecl, setVerbRequestBody } from './helpers'

test('shows echo verb form', async ({ page }) => {
  await navigateToDecl(page, 'echo', 'echo')

  await expect(page.getByText('CALL', { exact: true })).toBeVisible()
  await expect(page.locator('input#request-path')).toHaveValue('echo.echo')

  await expect(page.getByText('Body', { exact: true })).toBeVisible()
  await expect(page.getByText('Schema', { exact: true })).toBeVisible()
})

test('send echo request', async ({ page }) => {
  await navigateToDecl(page, 'echo', 'echo')

  await setVerbRequestBody(page, '{\n  "name": "wicket"\n}')

  await page.getByRole('button', { name: 'Send' }).click()

  const responseEditor = page.locator('#response-editor .cm-content[role="textbox"]')
  await expect(responseEditor).toBeVisible()

  const responseText = await responseEditor.textContent()
  const responseJson = JSON.parse(responseText?.trim() || '{}')

  const expectedStart = 'Hello, wicket!!! It is '
  expect(responseJson.message.startsWith(expectedStart)).toBe(true)
})
