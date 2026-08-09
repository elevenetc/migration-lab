import type {CellContext, CellDrawer} from 'active-grid';
import {lerpColor} from './colors';

const HOVER_LERP = 0.05;
const HOVER_RED = '#e63946';
const FOCUS_OUTLINE = '#98c1d9';
const SELECT_OUTLINE = '#e9c46a';

/** Base fill of a migration cell, handed to this drawer as `BackgroundData.base`. */
export const BASE_MIGRATION = '#243447';

function advanceHover(data: { hoverT: number }, cell: CellContext): number {
    const target = cell.hovered ? 1 : 0;
    data.hoverT += (target - data.hoverT) * HOVER_LERP;
    if (Math.abs(target - data.hoverT) < 0.01) data.hoverT = target;
    else cell.grid.invalidate();
    return data.hoverT;
}

export interface BackgroundData {
    base: string;
    hoverT: number;
}

/** Shared bottom layer: base fill with hover tint plus the focus and selection outlines. */
export const backgroundDrawer: CellDrawer<BackgroundData> = {
    draw(data, cell, ctx) {
        const {x, y, width, height} = cell.rect;
        ctx.fillStyle = lerpColor(data.base, HOVER_RED, advanceHover(data, cell));
        ctx.fillRect(x, y, width, height);
        if (cell.focused) {
            ctx.strokeStyle = FOCUS_OUTLINE;
            ctx.lineWidth = 4 / cell.scale;
            ctx.strokeRect(x, y, width, height);
        }
        if (cell.selected) {
            ctx.strokeStyle = SELECT_OUTLINE;
            ctx.lineWidth = 3 / cell.scale;
            ctx.strokeRect(x, y, width, height);
        }
    },
};
