import { useEffect, useRef } from 'react'
import { useMigrationStore } from '../store/migrationStore'
import { renderTimeline } from '../canvas/renderTimeline'

export function Timeline() {
  const { migrations, loading, error } = useMigrationStore()
  const canvasRef = useRef<HTMLCanvasElement>(null)

  // No dependency array: repaint on every render so data changes AND
  // Vite HMR edits (which re-render without changing `migrations`) redraw the canvas.
  useEffect(() => {
    if (canvasRef.current && migrations.length > 0) {
      renderTimeline(canvasRef.current, migrations)
    }
  })

  if (loading) {
    return <div style={{ padding: 20 }}>Loading migrations...</div>
  }

  if (error) {
    return <div style={{ padding: 20, color: 'red' }}>Error: {error}</div>
  }

  if (migrations.length === 0) {
    return <div style={{ padding: 20 }}>No migrations loaded</div>
  }

  return (
    <div style={{ width: '100%', height: '100%', overflow: 'auto', background: '#111' }}>
      <canvas ref={canvasRef} />
    </div>
  )
}
