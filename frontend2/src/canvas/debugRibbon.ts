import type {RibbonInfo} from './drawTransitionRibbon.ts'

export function debugRibbon(ctx: CanvasRenderingContext2D, ribbon: RibbonInfo): void {
    const {leftX, leftY, rightX, rightY, height} = ribbon
    const x = Math.min(leftX, rightX)
    const y = Math.min(leftY, rightY)
    const w = Math.abs(rightX - leftX)
    const h = Math.abs(rightY - leftY) + height

    ctx.strokeStyle = '#ff0000'
    ctx.beginPath()
    ctx.rect(x, y, w, h)
    ctx.stroke()
}
