import { useEffect, useMemo, useRef } from 'react'
import { useMigrationStore } from '../store/migrationStore'
import { buildMigrationCells } from '../grid/migrationCells'
import { buildMigrationGrid } from '../grid/buildMigrationGrid'
import '../grid/grid.css'

export function Timeline() {
  const { migrations, analysis, loading, error } = useMigrationStore()
  const canvasRef = useRef<HTMLCanvasElement>(null)

  const model = useMemo(
    () => buildMigrationCells(migrations, analysis?.warnings ?? []),
    [migrations, analysis],
  )
  const grid = useMemo(() => (model.cells.length > 0 ? buildMigrationGrid(model) : null), [model])

  // The grid owns its own resize observer and animation loop, so attaching is all there is to it.
  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas || !grid) return

    grid.attach(canvas)
    return () => grid.detach()
  }, [grid])

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
    <div style={{ position: 'relative', width: '100%', height: '100%', overflow: 'hidden', background: '#111' }}>
      <canvas ref={canvasRef} style={{ display: 'block', width: '100%', height: '100%' }} />
    </div>
  )
}
