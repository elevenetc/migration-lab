import { resolve } from 'node:path'
import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'
import { sentryVitePlugin } from '@sentry/vite-plugin'

export default defineConfig(({ mode }) => {
  const envDir = resolve(__dirname, '..')
  const env = loadEnv(mode, envDir, ['VITE_', 'SENTRY_'])
  const uploadSourceMaps = Boolean(env.SENTRY_AUTH_TOKEN && env.SENTRY_ORG && env.SENTRY_PROJECT)
  return {
    envDir,
    plugins: [react(), ...(uploadSourceMaps ? [sentryVitePlugin({
      org: env.SENTRY_ORG,
      project: env.SENTRY_PROJECT,
      authToken: env.SENTRY_AUTH_TOKEN,
      telemetry: false,
      release: { name: env.VITE_SENTRY_RELEASE || undefined },
      sourcemaps: { filesToDeleteAfterUpload: ['./dist/**/*.map'] },
    })] : [])],
    server: {
      port: 3000,
      proxy: {
        '/api': {
          target: process.env.VITE_API_TARGET ?? 'http://localhost:8080',
          changeOrigin: true
        }
      },
      watch: process.env.CHOKIDAR_USEPOLLING ? { usePolling: true } : undefined
    },
    build: {
      sourcemap: uploadSourceMaps ? 'hidden' : false,
      rollupOptions: {
        output: {
          manualChunks: undefined,
          inlineDynamicImports: true,
        }
      }
    }
  }
})
