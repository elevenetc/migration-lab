import {OperationLayoutInfo} from './layoutInfo.ts'
import {getOperationColor} from "./getOperationColor.ts";
import {getStringWidth} from "./getStringWidth.ts";
import {getStringHeight} from "./getStringHeight.ts";
import {getTableColor} from "./getTableColor.ts";

const FLAG_TITLE_CORNER_RADIUS = 4
const FLAG_TITLE_FLAG_NOTCH = 4

// Draws a single flag at x and returns the x where the next flag should start,
// so flags can be stacked left-to-right: foo< <bar< <hello<
export function drawFlag(
    title: string,
    first: boolean,
    x: number,
    op: OperationLayoutInfo,
    ctx: CanvasRenderingContext2D
): number {
    const w = getStringWidth(title, ctx) + 15
    const h = getStringHeight('A', ctx) + 6
    const y = op.y
    const r = FLAG_TITLE_CORNER_RADIUS
    const notch = FLAG_TITLE_FLAG_NOTCH

    const gradient = ctx.createLinearGradient(x, y, x + w, y)
    gradient.addColorStop(0, getTableColor(op.operation))
    gradient.addColorStop(1, getOperationColor(op.operation))
    ctx.fillStyle = gradient

    ctx.beginPath()
    ctx.moveTo(first ? x + r : x, y)
    ctx.lineTo(x + w, y)
    ctx.lineTo(x + w - notch, y + h / 2)
    ctx.lineTo(x + w, y + h)
    ctx.lineTo(first ? x + r : x, y + h)
    if (first) {
        // Flat left side with rounded corners.
        ctx.arcTo(x, y + h, x, y + h - r, r)
        ctx.lineTo(x, y + r)
        ctx.arcTo(x, y, x + r, y, r)
    } else {
        // Left notch mirrors the right one so it interlocks with the previous flag.
        ctx.lineTo(x - notch, y + h / 2)
    }
    ctx.closePath()
    ctx.fill()

    ctx.fillStyle = '#ffffff'
    ctx.font = '13px sans-serif'
    ctx.textAlign = 'center'
    ctx.textBaseline = 'middle'
    ctx.fillText(title, x + w / 2, y + h / 2)

    return x + w
}
