import type {CellDrawer} from 'active-grid';
import type {CellWarning} from './migration';

// World units, like the cell below it, so the icon zooms with the statement text.
const ICON = 26;
const MARGIN = 8;
/** Decor height that fits the icon with its margin; what callers pass to setHeader. */
export const WARNING_DECOR_HEIGHT = ICON + MARGIN;
// Below this on-screen size the icon is noise, so it is dropped.
const MIN_SCREEN_SIZE = 8;
const BORDER = '#e8a33d';
const MARK = '#1a1208';
const COUNT_FONT = `${ICON * 0.7}px ui-monospace, monospace`;

/** Static analysis warnings of one cell; the focus panel spells them out. */
export interface CellWarnings {
    warnings: CellWarning[];
}

function triangle(ctx: CanvasRenderingContext2D, x: number, y: number): void {
    ctx.beginPath();
    ctx.moveTo(x + ICON / 2, y);
    ctx.lineTo(x + ICON, y + ICON);
    ctx.lineTo(x, y + ICON);
    ctx.closePath();
    ctx.fillStyle = BORDER;
    ctx.fill();

    ctx.fillStyle = MARK;
    ctx.fillRect(x + ICON / 2 - 1.5, y + ICON * 0.35, 3, ICON * 0.35);
    ctx.fillRect(x + ICON / 2 - 1.5, y + ICON * 0.78, 3, 3);
}

/** Header decor of a migration cell: a right-aligned icon flagging the warnings raised on it. */
export const warningIconDrawer: CellDrawer<CellWarnings> = {
    draw(data, cell, ctx) {
        if (data.warnings.length === 0) return;
        if (ICON * cell.scale < MIN_SCREEN_SIZE) return;

        const x = cell.rect.x + cell.rect.width - ICON;
        const y = cell.rect.y + cell.rect.height - ICON;
        triangle(ctx, x, y);

        if (data.warnings.length > 1) {
            ctx.fillStyle = BORDER;
            ctx.font = COUNT_FONT;
            ctx.textAlign = 'right';
            ctx.textBaseline = 'bottom';
            ctx.fillText(String(data.warnings.length), x - MARGIN / 2, y + ICON);
        }
    },
};
