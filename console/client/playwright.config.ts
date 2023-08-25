import {defineConfig} from '@playwright/test';
export default defineConfig({
  testDir: '.playwright',
  use: {
    baseURL: 'http://localhost:6006/',
  },
});
