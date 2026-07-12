import { useEffect, useMemo, useRef } from 'react'
import { useMigrationStore } from '../store/migrationStore'
import { renderTimeline } from '../canvas/renderTimeline'
import { computeLayout } from '../canvas/layoutInfo'

export function Timeline() {
  const { migrations, analysis, loading, error } = useMigrationStore()
  const containerRef = useRef<HTMLDivElement>(null)
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const scrollRef = useRef({ x: 0, y: 0 })
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
      const maxX = Math.max(0, layout.width - width)
      const maxY = Math.max(0, layout.height - height)
      scrollRef.current.x = Math.min(Math.max(scrollRef.current.x, 0), maxX)
      scrollRef.current.y = Math.min(Math.max(scrollRef.current.y, 0), maxY)
      renderTimeline(canvas, layout, { width, height }, scrollRef.current)
    }
    paintRef.current = paint

    const onWheel = (e: WheelEvent) => {
      e.preventDefault()
      // Shift+wheel scrolls horizontally so mouse-wheel users can pan the x-axis.
      scrollRef.current.x += e.shiftKey ? e.deltaY : e.deltaX
      scrollRef.current.y += e.shiftKey ? 0 : e.deltaY
      paint()
    }

    container.addEventListener('wheel', onWheel, { passive: false })
    const ro = new ResizeObserver(paint)
    ro.observe(container)
    paint()

    return () => {
      container.removeEventListener('wheel', onWheel)
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
