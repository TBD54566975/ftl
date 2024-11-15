import { expect, test } from '@playwright/test'
import { navigateToDecl } from './helpers'

test('shows http ingress form', async ({ page }) => {
  await navigateToDecl(page, 'http', 'get')

  await expect(page.locator('#call-type')).toHaveText('GET')
  await expect(page.locator('input#request-path')).toHaveValue('http://localhost:8891/get/name')

  await expect(page.getByText('Schema', { exact: true })).toBeVisible()
})

test('send get request with path and query params', async ({ page }) => {
  await navigateToDecl(page, 'http', 'get')

  // Wait for the input to be stable with the initial value
  const pathInput = page.locator('input#request-path')
  await expect(pathInput).toHaveValue('http://localhost:8891/get/name', { timeout: 10000 })

  // Clear the input before filling to avoid concatenation issues
  await pathInput.clear()
  await pathInput.fill('http://localhost:8891/get/wicket')

  // Add a small wait to ensure the value is stable
  await page.waitForTimeout(100)

  // Click the Query Params tab
  await page.getByRole('button', { name: 'Query Params' }).click()

  // Fill out query parameters using the key/value inputs
  // Get the first pair of inputs in the form
  const keyInput = page.locator('input[placeholder="Key"]').first()
  const valueInput = page.locator('input[placeholder="Value"]').first()

  await keyInput.fill('age')
  await valueInput.fill('10')

  await page.getByRole('button', { name: 'Send' }).click()

  const responseEditor = page.locator('#response-editor .cm-content[role="textbox"]')
  await expect(responseEditor).toBeVisible()

  // Wait for valid JSON content
  await expect(async () => {
    const responseText = await responseEditor.textContent()
    expect(JSON.parse(responseText?.trim() || '{}')).toBeTruthy()
  }).toPass({ timeout: 5000 })

  const responseText = await responseEditor.textContent()
  const responseJson = JSON.parse(responseText?.trim() || '{}')

  expect(responseJson).toEqual({
    age: 10,
    name: 'wicket',
  })
})

test('send post request with body', async ({ page }) => {
  await navigateToDecl(page, 'http', 'post')

  await expect(page.locator('input#request-path')).toHaveValue('http://localhost:8891/post')

  const bodyEditor = page.locator('#body-editor .cm-content[contenteditable="true"]')
  await expect(bodyEditor).toBeVisible()
  await bodyEditor.fill('{\n  "age": 10,\n  "name": "wicket"\n}')

  await page.getByRole('button', { name: 'Send' }).click()

  const responseEditor = page.locator('#response-editor .cm-content[role="textbox"]')
  await expect(responseEditor).toBeVisible()

  const responseText = await responseEditor.textContent()
  const responseJson = JSON.parse(responseText?.trim() || '{}')

  expect(responseJson).toEqual({
    age: 10,
    name: 'wicket',
  })
})
