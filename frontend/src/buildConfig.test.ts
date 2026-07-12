import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

// vite.config.ts is loaded both to run the dev server and to produce the prod
// build, in containers/CI where only non-test dependencies may be present
// (e.g. a node_modules volume predating a devDependency, or a deps-pruned image).
// If it imports test-only tooling like vitest, the app fails to start/build.
// Regression guard for the frontend/frontend2 merge, where the vitest config
// was inlined into vite.config.ts and crashed the compose dev server.
describe('vite.config.ts', () => {
  const source = readFileSync(resolve(process.cwd(), 'vite.config.ts'), 'utf8')

  it('does not import test-only tooling', () => {
    expect(source).not.toMatch(/from\s+['"]vitest/)
    expect(source).not.toMatch(/require\(\s*['"]vitest/)
  })
})
