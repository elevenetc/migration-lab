import type {CellDrawer} from 'active-grid';
import {FG} from './colors';

const MIGRATION_SIZE = 20;
const MIGRATION_PAD = 16;
const MIGRATION_BAR = 6;
const MIGRATION_OP = MIGRATION_SIZE * 2;
const MIGRATION_OP_GAP = 12;
const MIGRATION_FONT = `${MIGRATION_SIZE}px ui-monospace, monospace`;
const MIGRATION_TAG_FONT = `${MIGRATION_SIZE * 0.6}px ui-monospace, monospace`;
const MIGRATION_RULE = 'rgba(224, 251, 252, 0.12)';

export type MigrationKind = 'create' | 'alter' | 'partition' | 'drop';

export const MIGRATION_KIND_COLOR: Record<MigrationKind, string> = {
    create: '#4caf82',
    alter: '#e8a33d',
    partition: '#7a9cd6',
    drop: '#e5646e',
};

/** One operation of a migration: what it does and what it touches. */
export interface CellOperation {
    action: string;
    target: string;
}

/** One static analysis warning raised on the cell's table by the cell's migration. */
export interface CellWarning {
    title: string;
    message: string;
}

export interface MigrationData {
    /** Flyway version of the migration this cell stands for; the header row spells it out. */
    version: string;
    /** Table the operations apply to. */
    table: string;
    /** Colors the bar and the action lines. */
    kind: MigrationKind;
    /** Applied top to bottom, in the order they appear in the migration. */
    operations: CellOperation[];
    /** SQL of every statement of the migration that touches this table; the focus panel prints it. */
    sql: string[];
    /** Static analysis warnings raised on this table by this migration; the focus panel explains them. */
    warnings: CellWarning[];
}

function operationTop(index: number): number {
    return MIGRATION_PAD + index * (MIGRATION_OP + MIGRATION_OP_GAP);
}

function widestText(texts: string[], ctx: CanvasRenderingContext2D): number {
    return texts.reduce((widest, text) => Math.max(widest, ctx.measureText(text).width), 0);
}

/** One migration applied to one table: kind bar and its operations stacked. The version is the header row's job. */
export const migrationDrawer: CellDrawer<MigrationData> = {
    layoutWidth(data, ctx) {
        ctx.font = MIGRATION_FONT;
        const target = widestText(data.operations.map(({target}) => target), ctx);
        ctx.font = MIGRATION_TAG_FONT;
        const action = widestText(data.operations.map(({action}) => action), ctx);
        return Math.max(target, action) + MIGRATION_PAD * 2 + MIGRATION_BAR;
    },

    layoutHeight(data) {
        return operationTop(Math.max(data.operations.length - 1, 0)) + MIGRATION_OP + MIGRATION_PAD;
    },

    draw(data, cell, ctx) {
        const {x, y, height} = cell.rect;
        const accent = MIGRATION_KIND_COLOR[data.kind];

        ctx.fillStyle = accent;
        ctx.fillRect(x, y, MIGRATION_BAR, height);

        const textX = x + MIGRATION_BAR + MIGRATION_PAD;
        ctx.textAlign = 'left';
        data.operations.forEach(({action, target}, index) => {
            const top = y + operationTop(index);

            if (index > 0) {
                ctx.fillStyle = MIGRATION_RULE;
                ctx.fillRect(textX, top - MIGRATION_OP_GAP / 2, cell.rect.width - MIGRATION_BAR - MIGRATION_PAD * 2, 1);
            }

            ctx.textBaseline = 'top';
            ctx.fillStyle = accent;
            ctx.font = MIGRATION_TAG_FONT;
            ctx.fillText(action, textX, top);

            ctx.textBaseline = 'bottom';
            ctx.fillStyle = FG;
            ctx.font = MIGRATION_FONT;
            ctx.fillText(target, textX, top + MIGRATION_OP);
        });
    },
};
