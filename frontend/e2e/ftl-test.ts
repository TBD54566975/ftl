import { test as base, expect } from '@playwright/test'

const ftlTest = base.extend<{
  // biome-ignore lint/suspicious/noExplicitAny: <explanation>
  page: any
}>({
  page: async ({ page }, use) => {
    await page.goto('http://localhost:8892')
    await page.waitForFunction(() => {
      const element = document.querySelector('#deployments-count')
      return element && element.textContent?.trim() === '2'
    })
    await use(page)
  },
})

export { ftlTest, expect }
