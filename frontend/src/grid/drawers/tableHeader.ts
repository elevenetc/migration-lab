import type {CellDrawer} from 'active-grid';
import {FG} from './colors';

const NAME_SIZE = 16;
const NAME_PAD = 14;
const RULE = 3;
const NAME_FONT = `${NAME_SIZE}px ui-monospace, monospace`;
const HEADER_FILL = 'rgba(13, 20, 28, 0.72)';
const ACCENT = '#98c1d9';

export interface TableHeaderData {
    table: string;
}

/** Row header: the table name against the right edge, lit while its row holds the selection. */
export const tableHeaderDrawer: CellDrawer<TableHeaderData> = {
    layoutWidth(data, ctx) {
        ctx.font = NAME_FONT;
        return ctx.measureText(data.table).width + NAME_PAD * 2;
    },

    layoutHeight() {
        return NAME_SIZE + NAME_PAD * 2;
    },

    draw(data, cell, ctx) {
        const {x, y, width, height} = cell.rect;
        const active = cell.grid.getSelected()?.row === cell.row;
        const rule = active ? RULE * 2 : RULE;

        ctx.fillStyle = HEADER_FILL;
        ctx.fillRect(x, y, width, height);
        ctx.fillStyle = ACCENT;
        ctx.fillRect(x + width - rule, y, rule, height);

        ctx.fillStyle = active ? ACCENT : FG;
        ctx.font = NAME_FONT;
        ctx.textAlign = 'right';
        ctx.textBaseline = 'middle';
        ctx.fillText(data.table, x + width - NAME_PAD, y + height / 2);
    },

    onClick() {
        return true;
    },
};
