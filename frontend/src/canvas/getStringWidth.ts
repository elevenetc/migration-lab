// Returns the width of text in pixels according to the current ctx font config.
export function getStringWidth(text: string, ctx: CanvasRenderingContext2D): number {
    return ctx.measureText(text).width
}
