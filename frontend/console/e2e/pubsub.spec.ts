import { expect, test } from '@playwright/test'
import { navigateToDecl, setVerbRequestBody } from './helpers'

test('shows pubsub verb form', async ({ page }) => {
  await navigateToDecl(page, 'pubsub', 'cookPizza')

  await expect(page.getByText('SUB', { exact: true })).toBeVisible()
  await expect(page.locator('input#request-path')).toHaveValue('pubsub.cookPizza')

  await expect(page.getByText('Body', { exact: true })).toBeVisible()
  await expect(page.getByText('Schema', { exact: true })).toBeVisible()
})

test('send pubsub request', async ({ page }) => {
  await navigateToDecl(page, 'pubsub', 'cookPizza')

  await setVerbRequestBody(page, '{\n  "customer": "wicket",\n  "id": 123,\n  "type": "cheese"\n}')

  await page.getByRole('button', { name: 'Send' }).click()

  const responseEditor = page.locator('#response-editor .cm-content[role="textbox"]')
  await expect(responseEditor).toBeVisible({ timeout: 10000 })

  const responseText = await responseEditor.textContent()
  const responseJson = JSON.parse(responseText?.trim() || '{}')

  expect(responseJson).toEqual({})
})
