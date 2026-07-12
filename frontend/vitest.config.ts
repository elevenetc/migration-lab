import { defineConfig } from 'vitest/config'

// Kept separate from vite.config.ts so the dev server and prod build never
// need vitest to load their config (see src/buildConfig.test.ts).
export default defineConfig({
  test: {
    environment: 'node',
    include: ['src/**/*.test.ts'],
  },
})
