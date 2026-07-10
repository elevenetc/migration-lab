// Curved ribbon flowing from the left edge down to the right edge: curved top
// and bottom edges, vertical left and right edges. Used for both rename and
// partition transitions.
export type RibbonInfo = {
    leftX: number
    leftY: number
    rightX: number
    rightY: number
    height: number
    leftColor: string
    rightColor: string
}

export function drawTransitionRibbon(ctx: CanvasRenderingContext2D, ribbon: RibbonInfo) {
    const { leftX, leftY, rightX, rightY, height, leftColor, rightColor } = ribbon
    const midX = (leftX + rightX) / 2

    ctx.beginPath()
    ctx.moveTo(leftX, leftY)
    ctx.bezierCurveTo(midX, leftY, midX, rightY, rightX, rightY)
    ctx.lineTo(rightX, rightY + height)
    ctx.bezierCurveTo(midX, rightY + height, midX, leftY + height, leftX, leftY + height)
    ctx.closePath()

    const gradient = ctx.createLinearGradient(leftX, leftY, rightX, rightY)
    gradient.addColorStop(0, leftColor)
    gradient.addColorStop(1, rightColor)
    ctx.fillStyle = gradient
    ctx.fill()
}
