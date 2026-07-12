// Curved ribbon flowing from the left edge down to the right edge: curved top
// and bottom edges, vertical left and right edges. Used for both rename and
// partition transitions.
import {debugRender} from "./drawOperation.ts";
import {debugRibbon} from "./debugRibbon.ts";

export type RibbonInfo = {
    leftX: number
    leftY: number
    rightX: number
    rightY: number
    height: number
    leftColor: string
    rightColor: string,
    init: boolean // if true, its should be rendered as first in the chain of migrations
}

const LEFT_CORNER_RADIUS = 4

export function drawTransitionRibbon(ctx: CanvasRenderingContext2D, ribbon: RibbonInfo) {
    const {leftX, leftY, rightX, rightY, height, leftColor, rightColor} = ribbon
    const midX = (leftX + rightX) / 2
    // The rightX end holds the first operation (create table); round its corners
    // when this ribbon starts a chain. hx steps horizontally toward the ribbon body.
    const r = ribbon.init ? LEFT_CORNER_RADIUS : 0
    const hx = Math.sign(leftX - rightX) * r

    ctx.beginPath()
    ctx.moveTo(leftX, leftY)
    ctx.bezierCurveTo(
        midX, leftY,
        midX, rightY,
        rightX + hx, rightY)
    ctx.arcTo(rightX, rightY, rightX, rightY + r, r)
    ctx.lineTo(rightX, rightY + height - r)
    ctx.arcTo(rightX, rightY + height, rightX + hx, rightY + height, r)
    ctx.bezierCurveTo(
        midX, rightY + height,
        midX, leftY + height,
        leftX, leftY + height)
    ctx.closePath()

    if (debugRender) {
        debugRibbon(ctx, ribbon)
    } else {
        const gradient = ctx.createLinearGradient(leftX, leftY, rightX, rightY)
        gradient.addColorStop(0, leftColor)
        gradient.addColorStop(1, rightColor)
        ctx.fillStyle = gradient
        ctx.fill()
    }
}
