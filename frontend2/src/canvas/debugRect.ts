import type {OperationLayoutInfo} from './layoutInfo.ts'

export function debugRect(ctx: CanvasRenderingContext2D, mig: OperationLayoutInfo): void {
    ctx.strokeStyle = '#ff0000'
    ctx.beginPath()
    ctx.rect(mig.x, mig.y, mig.w, mig.h)
    ctx.stroke()
}
