import { test as base, expect } from '@playwright/test'

const ftlTest = base.extend<{
  // biome-ignore lint/suspicious/noExplicitAny: <explanation>
  page: any
}>({
  page: async ({ page }, use) => {
    await page.goto('http://localhost:8892/modules')
    await page.waitForFunction(() => {
      const timeItem = document.querySelector('li#module-tree-item-time');
      const echoItem = document.querySelector('li#module-tree-item-echo');
      return timeItem !== null && echoItem !== null;
    });

    await use(page)
  },
})

export { ftlTest, expect }

