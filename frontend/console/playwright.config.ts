import { defineConfig, devices } from '@playwright/test'
import type { PlaywrightTestConfig } from '@playwright/test'

/**
 * See https://playwright.dev/docs/test-configuration.
 */
const config: PlaywrightTestConfig = {
  testDir: './e2e',
  globalSetup: './e2e/global-setup.ts',
  fullyParallel: true,
  /* Fail the build on CI if you accidentally left test.only in the source code. */
  forbidOnly: !!process.env.CI,
  /* Retry on CI only. */
  retries: process.env.CI ? 2 : 0,
  /* Opt out of parallel tests on CI. */
  workers: process.env.CI ? 1 : undefined,
  /* With additional example modules, it can take a bit of time for everything to start up. */
  timeout: 90 * 1000,
  reporter: 'html',
  use: {
    baseURL: 'http://localhost:8899',

    /* Collect trace when retrying the failed test. See https://playwright.dev/docs/trace-viewer */
    trace: 'on-first-retry',
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },

    {
      name: 'firefox',
      use: { ...devices['Desktop Firefox'] },
    },

    {
      name: 'webkit',
      use: { ...devices['Desktop Safari'] },
    },
  ],
  webServer: {
    command: process.env.CI ? 'ftl dev -j1' : 'ftl dev',
    url: 'http://localhost:8899',
    reuseExistingServer: !process.env.CI,
    /* If the test ends up needing to pull the postgres docker image, this can take a while. Give it a few minutes. */
    timeout: 180000,
  },
}

export default config
