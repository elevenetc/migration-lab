import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { ErrorBoundary, reactErrorHandler } from '@sentry/react'
import App from './App'
import { CrashFallback } from './components/CrashFallback'
import { initSentry } from './monitoring/initSentry'

initSentry()

createRoot(document.getElementById('root')!, {
  onUncaughtError: reactErrorHandler(),
  onRecoverableError: reactErrorHandler(),
  // ErrorBoundary reports caught errors; onCaughtError would report them twice.
}).render(
  <StrictMode>
    <ErrorBoundary fallback={<CrashFallback />}>
      <App />
    </ErrorBoundary>
  </StrictMode>,
)
