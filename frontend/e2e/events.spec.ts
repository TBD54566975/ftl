import { expect, ftlTest } from './ftl-test'

ftlTest('defaults to the events page', async ({ page }) => {
  const eventsNavItem = page.getByRole('link', { name: 'Events' })

  await expect(eventsNavItem).toHaveClass(/bg-indigo-600 text-white/)
})
