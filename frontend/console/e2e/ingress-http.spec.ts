import { expect, test } from '@playwright/test'
import { navigateToDecl } from './helpers'

test('shows http ingress form', async ({ page }) => {
  await navigateToDecl(page, 'http', 'get')

  await expect(page.locator('#call-type')).toHaveText('GET')
  await expect(page.locator('input#request-path')).toHaveValue('http://localhost:8891/get/name')

  await expect(page.getByText('Body', { exact: true })).toBeVisible()
  await expect(page.getByText('Verb Schema', { exact: true })).toBeVisible()
  await expect(page.getByText('JSONSchema', { exact: true })).toBeVisible()
})

test('send get request with path and query params', async ({ page }) => {
  await navigateToDecl(page, 'http', 'get')

  // Check the initial value of the path input
  const pathInput = page.locator('input#request-path')
  expect(pathInput).toHaveValue('http://localhost:8891/get/name')

  // Update the path input to test path and query params
  await pathInput.fill('http://localhost:8891/get/wicket?age=10')
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
