import { useEffect, useMemo, useRef } from 'react'
import { useMigrationStore } from '../store/migrationStore'
import { renderTimeline } from '../canvas/renderTimeline'
import { computeLayout } from '../canvas/layoutInfo'
import { computeFitScale } from '../canvas/fitScale'

export function Timeline() {
  const { migrations, analysis, loading, error } = useMigrationStore()
  const containerRef = useRef<HTMLDivElement>(null)
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const paintRef = useRef<() => void>(() => {})

  const layout = useMemo(
    () => (migrations.length > 0 ? computeLayout(migrations, analysis?.warnings ?? []) : null),
    [migrations, analysis],
  )

  useEffect(() => {
    const container = containerRef.current
    const canvas = canvasRef.current
    if (!container || !canvas || !layout) return

    const paint = () => {
      const { width, height } = container.getBoundingClientRect()
      const scale = computeFitScale({ width, height }, layout)
      renderTimeline(canvas, layout, { width, height }, scale)
    }
    paintRef.current = paint

    const ro = new ResizeObserver(paint)
    ro.observe(container)
    paint()

    return () => {
      ro.disconnect()
    }
  }, [layout])

  // No dependency array: repaint on every render so Vite HMR edits (which
  // re-render without changing `layout`) redraw the canvas.
  useEffect(() => {
    paintRef.current()
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
    <div ref={containerRef} style={{ width: '100%', height: '100%', overflow: 'hidden', background: '#111' }}>
      <canvas ref={canvasRef} style={{ display: 'block' }} />
    </div>
  )
}
