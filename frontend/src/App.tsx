import { useEffect } from 'react'
import { Timeline } from './components/Timeline'
import { useMigrationStore } from './store/migrationStore'

function App() {
  const loadMigrations = useMigrationStore((state) => state.loadMigrations)
  const runMigrations = useMigrationStore((state) => state.runMigrations)
  const running = useMigrationStore((state) => state.running)
  const runResult = useMigrationStore((state) => state.runResult)

  useEffect(() => {
    loadMigrations()
  }, [loadMigrations])

  return (
    <div style={{ width: '100vw', height: '100vh', display: 'flex', flexDirection: 'column' }}>
      <header style={{ padding: '16px', borderBottom: '1px solid #333', background: '#1a1a1a' }}>
        <h1 style={{ margin: 0, fontSize: '20px', color: '#fff' }}>Migration Timeline</h1>
        <div style={{ marginTop: '8px', display: 'flex', alignItems: 'center', gap: '12px' }}>
          <button
            style={{
              padding: '8px 16px',
              background: '#4a5568',
              color: 'white',
              border: 'none',
              borderRadius: '4px',
              cursor: 'pointer',
            }}
            disabled
          >
            Select migrations (coming soon)
          </button>
          <button
            onClick={runMigrations}
            disabled={running}
            style={{
              padding: '8px 16px',
              background: running ? '#4a5568' : '#22c55e',
              color: 'white',
              border: 'none',
              borderRadius: '4px',
              cursor: running ? 'not-allowed' : 'pointer',
            }}
          >
            {running ? 'Running...' : 'Run migrations'}
          </button>
          {runResult && (
            <span style={{ color: runResult.success ? '#22c55e' : '#ef4444', fontSize: '14px' }}>
              {runResult.success
                ? `Applied ${runResult.migrationsApplied} migrations`
                : runResult.message}
            </span>
          )}
        </div>
      </header>
      <main style={{ flex: 1 }}>
        <Timeline />
      </main>
    </div>
  )
}

export default App
