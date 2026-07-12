import {OperationLayoutInfo, TRANSITION_TAG_SHIFT} from './layoutInfo.ts'
import {getOperationTitle} from "./getOperationTitle.ts";
import {debugRect} from "./debugRect.ts";
import {getTableColor} from "./getTableColor.ts";
import {drawFlag} from "./drawFlag.ts";

export const debugRender = false

export function drawOperation(
    ctx: CanvasRenderingContext2D,
    op: OperationLayoutInfo,
    prevOp: OperationLayoutInfo | null,
    nextOp: OperationLayoutInfo | null,
    drawGradient: boolean
): void {
    if (debugRender) {
        debugRect(ctx, op)
    } else {
        if (drawGradient) drawBackground(prevOp, nextOp, op, ctx);
    }
    drawFlag(getOperationTitle(op.operation), true, op, ctx);
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
