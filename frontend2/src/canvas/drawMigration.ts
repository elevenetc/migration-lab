import type {OperationLayoutInfo} from './layoutInfo.ts'
import {getOperationTitle} from "./getOperationTitle.ts";
import {getOperationColor} from "./getOperationColor.ts";

// Draws a single migration as a rounded rectangle with its name centered.
export function drawMigration(
    ctx: CanvasRenderingContext2D,
    mig: OperationLayoutInfo,
    nextMig: OperationLayoutInfo | null,
    drawGradient: boolean
): void {
    drawBackground(nextMig, mig, ctx, drawGradient);
    drawTitle(ctx, mig);
}

function drawTitle(ctx: CanvasRenderingContext2D, op: OperationLayoutInfo) {
    ctx.fillStyle = '#ffffff'
    ctx.font = '13px sans-serif'
    ctx.textAlign = 'center'
    ctx.textBaseline = 'middle'

    ctx.fillText(getOperationTitle(op.operation), op.x + op.w / 2, op.y + op.h / 2)
}

function drawBackground(nextMig: OperationLayoutInfo | null, mig: OperationLayoutInfo, ctx: CanvasRenderingContext2D, drawGradient: boolean) {
    const migW = nextMig == null ? mig.w : nextMig.x - mig.x;

    const gradient = ctx.createLinearGradient(mig.x, mig.y, mig.x + migW, mig.y)
    let operationColor = getOperationColor(mig.operation);
    gradient.addColorStop(0, operationColor)

    if (drawGradient) {
        gradient.addColorStop(1, '#00000000')
    } else {
        gradient.addColorStop(0, operationColor)
    }


    ctx.fillStyle = gradient
    ctx.beginPath()
    ctx.rect(mig.x, mig.y, migW, mig.h)
    ctx.fill()
}

