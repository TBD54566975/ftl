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

  // Some decls are hidden by default because they are not exported, so click
  // the toggle to make them visible.
  await page.locator('#hide-exported').click()

  // Navigate to the decl page
  await page.locator(`a#decl-${declName}`).click()
  await expect(page).toHaveURL(new RegExp(`/modules/${moduleName}/verb/${declName}`))
}

export async function pressShortcut(page: Page, key: string) {
  // Get the platform-specific modifier key
  const isMac = await page.evaluate(() => navigator.userAgent.includes('Mac'))
  const modifier = isMac ? 'Meta' : 'Control'

  await page.keyboard.down(modifier)
  await page.keyboard.press(key)
  await page.keyboard.up(modifier)
}
