// Uniform scale that fits the whole layout into the viewport, clamped at 1
// so small timelines render at natural size.
export function computeFitScale(
    viewport: { width: number; height: number },
    layout: { width: number; height: number },
): number {
    return Math.min(1, viewport.width / layout.width, viewport.height / layout.height)
}
