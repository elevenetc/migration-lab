import type {CellDrawer} from 'active-grid';
import {FG} from './colors';
import {MIGRATION_KIND_COLOR, type MigrationKind} from './migration';

const NAME_SIZE = 14;
const NAME_PAD = 12;
const RULE = 3;
const NAME_FONT = `${NAME_SIZE}px ui-monospace, monospace`;
const HEADER_FILL = 'rgba(13, 20, 28, 0.72)';

export interface MigrationHeaderData {
    name: string;
    kind: MigrationKind;
}

/** Column header: the migration version over a rule in its kind's color, lit while its column holds the selection. */
export const migrationHeaderDrawer: CellDrawer<MigrationHeaderData> = {
    layoutWidth(data, ctx) {
        ctx.font = NAME_FONT;
        return ctx.measureText(data.name).width + NAME_PAD * 2;
    },

    layoutHeight() {
        return NAME_SIZE + NAME_PAD * 2;
    },

    draw(data, cell, ctx) {
        const {x, y, width, height} = cell.rect;
        const accent = MIGRATION_KIND_COLOR[data.kind];
        const active = cell.grid.getSelected()?.col === cell.col;
        const rule = active ? RULE * 2 : RULE;

        ctx.fillStyle = HEADER_FILL;
        ctx.fillRect(x, y, width, height);
        ctx.fillStyle = accent;
        ctx.fillRect(x, y + height - rule, width, rule);

        ctx.fillStyle = active ? accent : FG;
        ctx.font = NAME_FONT;
        ctx.textAlign = 'center';
        ctx.textBaseline = 'middle';
        ctx.fillText(data.name, x + width / 2, y + height / 2);
    },

    onClick() {
        return true;
    },
};
