import { expect, ftlTest } from './ftl-test'

ftlTest('shows command palette results', async ({ page }) => {
  await page.goto('http://localhost:8892')

  await page.click('#command-palette-search')
  await page.fill('#search-input', 'echo')

  // Command palette should be showing the echo parts
  await expect(page.getByText('echo.EchoRequest')).toBeVisible()
  await expect(page.getByText('echo.EchoReponse')).toBeVisible()
  await expect(page.getByText('echo.echo')).toBeVisible()
})
