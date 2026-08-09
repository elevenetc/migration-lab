import {type CellContext, type CellDrawer, type Rect, rectContains} from 'active-grid';
import {FG} from './colors';

// World units, like every other layer in a migration cell, so the button zooms with the operation text.
const WIDTH = 88;
const HEIGHT = 26;
const MARGIN = 8;
/** Decor height that fits the button with its margin; what callers pass to setFooter. */
export const FOCUS_BUTTON_DECOR_HEIGHT = HEIGHT + MARGIN * 2;
const RADIUS = 4;
const FONT_SIZE = 13;
// Below this on-screen height the label is noise, so the button is dropped.
const MIN_SCREEN_HEIGHT = 8;
const FILL = 'rgba(20, 32, 42, 0.85)';
const BORDER = '#98c1d9';

const LABEL = 'focus-in';

function visible(cell: CellContext): boolean {
    return cell.hovered && HEIGHT * cell.scale >= MIN_SCREEN_HEIGHT;
}

/** Right-aligned inside the decor the button lives in. */
function buttonRect(cell: CellContext): Rect {
    return {
        x: cell.rect.x + cell.rect.width - WIDTH - MARGIN,
        y: cell.rect.y + (cell.rect.height - HEIGHT) / 2,
        width: WIDTH,
        height: HEIGHT,
    };
}

function hitsButton(cell: CellContext): boolean {
    return visible(cell) && cell.pointer !== null && rectContains(buttonRect(cell), cell.pointer);
}

/** Footer decor of a migration cell: zooms the grid into the cell. */
export const focusButtonDrawer: CellDrawer<null> = {
    draw(_data, cell, ctx) {
        if (!visible(cell)) return;
        const {x, y, width, height} = buttonRect(cell);
        const hot = hitsButton(cell);

        ctx.beginPath();
        ctx.roundRect(x, y, width, height, RADIUS);
        ctx.fillStyle = FILL;
        ctx.fill();
        ctx.strokeStyle = BORDER;
        ctx.lineWidth = (hot ? 2 : 1) / cell.scale;
        ctx.stroke();

        ctx.fillStyle = hot ? BORDER : FG;
        ctx.font = `${FONT_SIZE}px system-ui, sans-serif`;
        ctx.textAlign = 'center';
        ctx.textBaseline = 'middle';
        ctx.fillText(LABEL, x + width / 2, y + height / 2);
    },

    onClick(_data, cell) {
        if (!hitsButton(cell)) return;
        cell.grid.focusCell(cell.row, cell.col);
        return true;
    },
};
