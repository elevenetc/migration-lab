import {type CellContext, type Rect, rectContains} from 'active-grid';
import {FG} from './colors';

// World units, like every other layer in a migration cell, so the buttons zoom with the operation text.
const WIDTH = 88;
const HEIGHT = 26;
const MARGIN = 8;
const GAP = 6;
const RADIUS = 4;
const FONT_SIZE = 13;
// Below this on-screen height the label is noise, so the buttons are dropped.
const MIN_SCREEN_HEIGHT = 8;
const FILL = 'rgba(20, 32, 42, 0.85)';

/** Decor height that fits the given number of stacked buttons; what callers pass to setFooter. */
export function buttonsDecorHeight(count: number): number {
    return MARGIN * 2 + count * HEIGHT + Math.max(count - 1, 0) * GAP;
}

/** Buttons belong to the cell under the pointer, and only while they are big enough to read. */
export function buttonVisible(cell: CellContext): boolean {
    return cell.hovered && HEIGHT * cell.scale >= MIN_SCREEN_HEIGHT;
}

/** Right-aligned inside the decor the buttons live in, stacked top-down by slot. */
export function buttonRect(cell: CellContext, slot: number): Rect {
    return {
        x: cell.rect.x + cell.rect.width - WIDTH - MARGIN,
        y: cell.rect.y + MARGIN + slot * (HEIGHT + GAP),
        width: WIDTH,
        height: HEIGHT,
    };
}

export function buttonHit(cell: CellContext, slot: number): boolean {
    return buttonVisible(cell) && cell.pointer !== null && rectContains(buttonRect(cell, slot), cell.pointer);
}

/** Draws one slot of the footer as a labelled button, outlined in its own color. */
export function drawButton(
    cell: CellContext,
    ctx: CanvasRenderingContext2D,
    slot: number,
    label: string,
    color: string,
): void {
    if (!buttonVisible(cell)) return;

    const {x, y, width, height} = buttonRect(cell, slot);
    const hot = buttonHit(cell, slot);

    ctx.beginPath();
    ctx.roundRect(x, y, width, height, RADIUS);
    ctx.fillStyle = FILL;
    ctx.fill();
    ctx.strokeStyle = color;
    ctx.lineWidth = (hot ? 2 : 1) / cell.scale;
    ctx.stroke();

    ctx.fillStyle = hot ? color : FG;
    ctx.font = `${FONT_SIZE}px system-ui, sans-serif`;
    ctx.textAlign = 'center';
    ctx.textBaseline = 'middle';
    ctx.fillText(label, x + width / 2, y + height / 2);
}
