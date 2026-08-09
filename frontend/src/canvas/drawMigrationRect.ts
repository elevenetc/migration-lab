import {MigrationLayoutInfo} from './layoutInfo.ts'

const CORNER_RADIUS = 4

// Basic outline around all operations of one migration, painted behind them.
// `scale` is the world-to-screen fit scale; the stroke stays 1 screen pixel wide.
export function drawMigrationRect(ctx: CanvasRenderingContext2D, bounds: MigrationLayoutInfo, scale: number): void {
    ctx.strokeStyle = 'rgba(255, 0, 0, 1)'
    ctx.lineWidth = 1 / scale
    ctx.beginPath()
    ctx.roundRect(bounds.x, bounds.y, bounds.w, bounds.h, CORNER_RADIUS)
    ctx.stroke()
}
