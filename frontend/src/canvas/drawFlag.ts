import {OperationLayoutInfo} from './layoutInfo.ts'
import {getOperationColor} from "./getOperationColor.ts";
import {getStringWidth} from "./getStringWidth.ts";
import {getStringHeight} from "./getStringHeight.ts";
import {getTableColor} from "./getTableColor.ts";

const FLAG_TITLE_CORNER_RADIUS = 4
const FLAG_TITLE_FLAG_NOTCH = 4

export function drawFlag(
    title: string,
    first: boolean,
    op: OperationLayoutInfo,
    ctx: CanvasRenderingContext2D
) {
    let titleWidth = getStringWidth(title, ctx) + 15
    let titleHeight = getStringHeight('A', ctx) + 6

    const x = op.x
    const y = op.y
    const w = titleWidth
    const h = titleHeight
    const r = FLAG_TITLE_CORNER_RADIUS
    const notch = FLAG_TITLE_FLAG_NOTCH

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
