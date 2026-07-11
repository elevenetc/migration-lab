import {OperationLayoutInfo, TRANSITION_TAG_SHIFT} from './layoutInfo.ts'
import {getOperationTitle} from "./getOperationTitle.ts";
import {getOperationColor} from "./getOperationColor.ts";
import {getStringWidth} from "./getStringWidth.ts";
import {getStringHeight} from "./getStringHeight.ts";
import {debugRect} from "./debugRect.ts";
import {getTableColor} from "./getTableColor.ts";

export const debugRender = false

export function drawOperation(
    ctx: CanvasRenderingContext2D,
    operation: OperationLayoutInfo,
    prevOperation: OperationLayoutInfo | null,
    nextOperation: OperationLayoutInfo | null,
    drawGradient: boolean
): void {
    if (debugRender) {
        debugRect(ctx, operation)
    } else {
        if (drawGradient) drawBackground(prevOperation, nextOperation, operation, ctx);
    }

    drawTitle(ctx, operation);
}

const TITLE_CORNER_RADIUS = 4
const TITLE_FLAG_NOTCH = 4

function drawTitle(ctx: CanvasRenderingContext2D, op: OperationLayoutInfo) {

    let title = getOperationTitle(op.operation)
    let titleWidth = getStringWidth(title, ctx) + 15
    let titleHeight = getStringHeight('A', ctx) + 6

    const x = op.x
    const y = op.y
    const w = titleWidth
    const h = titleHeight
    const r = TITLE_CORNER_RADIUS
    const notch = TITLE_FLAG_NOTCH

    const gradient = ctx.createLinearGradient(x, y, x + w, y)
    gradient.addColorStop(0, getTableColor(op.operation))
    gradient.addColorStop(1, getOperationColor(op.operation))
    ctx.fillStyle = gradient

    ctx.beginPath()
    ctx.moveTo(x + r, y)
    ctx.lineTo(x + w, y)
    ctx.lineTo(x + w - notch, y + h / 2)
    ctx.lineTo(x + w, y + h)
    ctx.lineTo(x + r, y + h)
    ctx.arcTo(x, y + h, x, y + h - r, r)
    ctx.lineTo(x, y + r)
    ctx.arcTo(x, y, x + r, y, r)
    ctx.closePath()
    ctx.fill()

    ctx.beginPath()
    ctx.moveTo(x + 1, y + 5)
    ctx.lineTo(x + 1, y + 30)
    ctx.strokeStyle = getTableColor(op.operation)
    ctx.lineWidth = 2
    ctx.stroke()

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

const BACKGROUND_CORNER_RADIUS = 4

function drawBackground(
    prevOp: OperationLayoutInfo | null,
    nextOp: OperationLayoutInfo | null,
    currentOp: OperationLayoutInfo,
    ctx: CanvasRenderingContext2D
) {
    const w = nextOp == null ? currentOp.w : nextOp.x - currentOp.x;

    let operationColor = getTableColor(currentOp.operation)

    // First operation for the table gets rounded corners on the left side.
    const r = isFirstOperation(prevOp, currentOp) ? BACKGROUND_CORNER_RADIUS : 0
    const corners: [number, number, number, number] = [r, 0, 0, r]

    if (nextOp == null) {
        const gradient = ctx.createLinearGradient(
            currentOp.x, currentOp.y,
            currentOp.x + w, currentOp.y
        )

        gradient.addColorStop(0, operationColor)
        gradient.addColorStop(1, '#00000000')

        ctx.fillStyle = gradient
    } else {
        ctx.fillStyle = operationColor
    }

    ctx.beginPath()
    ctx.roundRect(currentOp.x, currentOp.y + TRANSITION_TAG_SHIFT, w, currentOp.h - TRANSITION_TAG_SHIFT, corners)
    ctx.fill()
}

function isFirstOperation(prevOp: OperationLayoutInfo | null, currentOp: OperationLayoutInfo) {
    if (currentOp.operation.type == 'CREATE_TABLE') {
        if (currentOp.operation.partitionOf) return false
    }
    return prevOp == null && currentOp.operation.type != 'RENAME_TABLE';
}
