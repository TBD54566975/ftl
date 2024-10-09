import { expect, test } from '@playwright/test'
import { navigateToDecl } from './helpers'

test('shows echo verb form', async ({ page }) => {
  await navigateToDecl(page, 'echo', 'echo')

  await expect(page.getByText('CALL', { exact: true })).toBeVisible()
  await expect(page.locator('input#request-path')).toHaveValue('echo.echo')

  await expect(page.getByText('Body', { exact: true })).toBeVisible()
  await expect(page.getByText('Verb Schema', { exact: true })).toBeVisible()
  await expect(page.getByText('JSONSchema', { exact: true })).toBeVisible()
})

test('send echo request', async ({ page }) => {
  await navigateToDecl(page, 'echo', 'echo')

  // Check the initial value of the path input
  const pathInput = page.locator('#request-path')
  await expect(pathInput).toBeVisible()
  const currentValue = await pathInput.inputValue()
  expect(currentValue).toBe('echo.echo')

  const bodyEditor = page.locator('#body-editor .cm-content[contenteditable="true"]')
  await expect(bodyEditor).toBeVisible()
  await bodyEditor.fill('{\n  "name": "wicket"\n}')

  await page.getByRole('button', { name: 'Send' }).click()

  const responseEditor = page.locator('#response-editor .cm-content[role="textbox"]')
  await expect(responseEditor).toBeVisible()

  const responseText = await responseEditor.textContent()
  const responseJson = JSON.parse(responseText?.trim() || '{}')

  const expectedStart = 'Hello, wicket!!! It is '
  expect(responseJson.message.startsWith(expectedStart)).toBe(true)
})
