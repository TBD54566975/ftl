import { expect, test } from '@playwright/test'
import { navigateToModule } from './helpers'

test('shows verbs for deployment', async ({ page }) => {
  await navigateToModule(page, 'echo')

  await expect(page.getByText('module echo {')).toBeVisible()
})
