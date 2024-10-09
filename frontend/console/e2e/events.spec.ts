import { expect, test } from '@playwright/test'

test('defaults to the events page', async ({ page }) => {
  await page.goto('http://localhost:8892')
  const eventsNavItem = page.getByRole('link', { name: 'Events' })

  await expect(eventsNavItem).toHaveClass(/bg-indigo-700 text-white/)
})
