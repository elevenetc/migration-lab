// Returns the height of text in pixels according to the current ctx font config.
export function getStringHeight(text: string, ctx: CanvasRenderingContext2D): number {
    const metrics = ctx.measureText(text)
    return metrics.actualBoundingBoxAscent + metrics.actualBoundingBoxDescent
}
