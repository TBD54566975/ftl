/* eslint-disable no-console */
const { chromium } = require('playwright')
const fs = require('fs/promises')
const path = require('path')
const process = require('process')
const fg = require('fast-glob')

const template = ({ id, name, kind }) => {

  const importPlaywright = "import { test, expect } from '@playwright/test'"

  const importAxeCore =  "import AxeBuilder from '@axe-core/playwright'"

  const AccessibilityTest = `test('Accessibility check ${kind}: ${name} story', async ({ page }) => {
  await page.goto('./iframe.html?id=${id}');
  const results = await new AxeBuilder({ page }).options({}).include('#storybook-root').analyze();
  expect(results.violations).toHaveLength(0);
})`

  const VisualComparisonTest = `test('Renders ${kind}: ${name} story', async ({ page }) => {
  await page.goto('./iframe.html?id=${id}');
  await page.waitForTimeout(800);
  expect(await page.screenshot()).toMatchSnapshot('${id}.png');
});`

  return [
    importPlaywright,
    importAxeCore,
    ' ',
    AccessibilityTest,
    VisualComparisonTest,
  ].filter(Boolean).join('\n')
}

// Build playwright tests
const build = async url => {
  const browser = await chromium.launch()
  const context = await browser.newContext()
  const page = await context.newPage()

  await page.goto(url, {
    waitUntil: 'load',
  })

  try {
    storyData = await page.evaluate(async () => {
      await await window.__STORYBOOK_CLIENT_API__.storyStore.cacheAllCSFFiles()
      return  Object.values(window.__STORYBOOK_STORY_STORE__.getStoriesJsonData().stories)
    })
  } catch (err) {
    console.error(err)
    process.exit(1)
  }
  // Make test directory if it does not exist
  const testDirectory = path.resolve(__dirname, `../.playwright`)
  await fs.mkdir(testDirectory, { recursive: true })

  // Put story id's into set for cleanup
  const storyDataIds = new Set(storyData.map(({ id }) => id))
  // Glob old tests
  const oldTests = await fg([ path.resolve(__dirname, '../.playwright/*.spec.ts') ], { dot: true })
  // Cleanup test no longer needed
  oldTests.length && await Promise.all(
    oldTests.map(async filePath => {
      const fileBaseName = path.basename(filePath, '.spec.ts')
      // Check if the id of the old test file is in the storyDataIds Set
      if (!storyDataIds.has(fileBaseName)) {
        await fs.rm(filePath)
        console.log(`Removed old test file: "${filePath}"`)
      }
    })
  )

  await Promise.all(
    storyData.map(async ({ id, name, kind }) => {
      const filePath = `${testDirectory}/${id}.spec.ts`

      // Check if the test file exists
      try {
        await fs.access(filePath)
        console.log(`Test file "${filePath}" already exists, skipping...`)
      } catch (error) {
        // If the test file does not exist, create it
        await fs.writeFile(filePath, template({ id, name, kind }))
      }
    })
  )

  await browser.close()
}



(async function() {
  const baseURL = 'http://localhost:6006'
  const url = `${baseURL}/iframe.html`
  try {
    await build(url)
    // eslint-disable-next-line no-console
    console.log('Process completed successfully.')
    process.exit(0)
  } catch (error) {
    console.error('An error occurred during the build process: ', error)
    process.exit(1)
  }
})()
