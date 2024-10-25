import { type Page, expect } from '@playwright/test'

export async function navigateToModule(page: Page, moduleName: string) {
  await page.goto('http://localhost:8892/modules')
  await page.getByRole('link', { name: 'Modules' }).click()

  // Navigate to the module page
  await page.locator(`#module-${moduleName}-view-icon`).click()
  await expect(page).toHaveURL(new RegExp(`/modules/${moduleName}`))

  // Expand the module tree group
  await page.locator(`#module-${moduleName}-tree-group`).click()
}

export async function navigateToDecl(page: Page, moduleName: string, declName: string) {
  await navigateToModule(page, moduleName)

  // Navigate to the decl page
  await page.locator(`a#decl-${declName}`).click()
  await expect(page).toHaveURL(new RegExp(`/modules/${moduleName}/verb/${declName}`))
}
