import { type FullConfig, chromium } from '@playwright/test'

const globalSetup = async (config: FullConfig) => {
  console.log('Waiting for server to be available...')

  const browser = await chromium.launch()
  const context = await browser.newContext()
  const page = await context.newPage()
  await page.goto('http://localhost:8892/modules')

  console.log('Waiting for modules to load...')
  const moduleNames = ['time', 'echo', 'cron', 'fsm', 'http', 'pubsub']
  await page.waitForFunction(
    (modules) => {
      const loadedModules = modules.filter((module) => document.querySelector(`li#module-tree-module-${module}`) !== null)
      console.log('Loaded modules:', loadedModules.join(', '))
      return loadedModules.length === modules.length
    },
    moduleNames,
    { timeout: 120000 },
  )

  console.log('Modules loaded!')

  await browser.close()
}

export default globalSetup
