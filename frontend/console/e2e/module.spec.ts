import { expect, ftlTest } from './ftl-test'
import { navigateToModule } from './helpers'

ftlTest('shows verbs for deployment', async ({ page }) => {
  await navigateToModule(page, 'echo')

  await expect(page.getByText('module echo {')).toBeVisible()
})
