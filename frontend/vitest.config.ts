import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: 'src/setupTests.ts',
    exclude: [
      'e2e/**/*',
      'node_modules/**/*',
    ],
  },
});
