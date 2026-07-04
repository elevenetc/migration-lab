import { useEffect } from 'react'
import { Timeline } from './components/Timeline'
import { useMigrationStore } from './store/migrationStore'

function App() {
  const loadMigrations = useMigrationStore((state) => state.loadMigrations)

  useEffect(() => {
    loadMigrations()
  }, [loadMigrations])

  return (
    <div style={{ width: '100vw', height: '100vh', display: 'flex', flexDirection: 'column' }}>
      <header style={{ padding: '16px', borderBottom: '1px solid #333', background: '#1a1a1a' }}>
        <h1 style={{ margin: 0, fontSize: '20px', color: '#fff' }}>Migration Timeline</h1>
        <button
          style={{
            marginTop: '8px',
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
      </header>
      <main style={{ flex: 1 }}>
        <Timeline />
      </main>
    </div>
  )
}

export default App
