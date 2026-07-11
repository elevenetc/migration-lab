import {OperationLayoutInfo, TRANSITION_TAG_SHIFT} from './layoutInfo.ts'
import {getOperationTitle} from "./getOperationTitle.ts";
import {getOperationColor} from "./getOperationColor.ts";
import {getStringWidth} from "./getStringWidth.ts";
import {getStringHeight} from "./getStringHeight.ts";
import {debugRect} from "./debugRect.ts";

export const debugRender = false

// Draws a single migration as a rounded rectangle with its name centered.
export function drawMigration(
    ctx: CanvasRenderingContext2D,
    mig: OperationLayoutInfo,
    nextMig: OperationLayoutInfo | null,
    drawGradient: boolean
): void {
    if (debugRender) {
        debugRect(ctx, mig)
    } else {
        if (drawGradient) drawBackground(nextMig, mig, ctx);
    }

    drawTitle(ctx, mig);
}

function drawTitle(ctx: CanvasRenderingContext2D, op: OperationLayoutInfo) {

    let title = getOperationTitle(op.operation)
    let titleWidth = getStringWidth(title, ctx) + 15
    let titleHeight = getStringHeight('A', ctx) + 6

    ctx.fillStyle = getOperationColor(op.operation)
    ctx.beginPath()
    ctx.roundRect(op.x, op.y, titleWidth, titleHeight, 3)
    ctx.fill()

    ctx.fillStyle = '#ffffff'
    ctx.font = '13px sans-serif'
    ctx.textAlign = 'center'
    ctx.textBaseline = 'middle'

    ctx.fillText(
        title,
        op.x + titleWidth / 2,
        op.y + titleHeight / 2
    )
}

function drawBackground(
    nextMig: OperationLayoutInfo | null,
    mig: OperationLayoutInfo,
    ctx: CanvasRenderingContext2D) {
    const migW = nextMig == null ? mig.w : nextMig.x - mig.x;

    const gradient = ctx.createLinearGradient(
        mig.x, mig.y,
        mig.x + migW, mig.y
    )
    let operationColor = getOperationColor(mig.operation);
    gradient.addColorStop(0, operationColor)
    gradient.addColorStop(1, '#00000000')


    ctx.fillStyle = gradient
    ctx.beginPath()
    ctx.rect(mig.x, mig.y + TRANSITION_TAG_SHIFT, migW, mig.h - TRANSITION_TAG_SHIFT)
    ctx.fill()
}

