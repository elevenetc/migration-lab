import {type CellDrawer} from 'active-grid';
import {buttonHit, drawButton} from './cellButton';

const SLOT = 1;
const LABEL = 'runtime';
const COLOR = '#4caf82';

/** Which migration the button measures, and what to hand it to. */
export interface RuntimeButtonData {
    migrationId: string;
    run: (migrationId: string) => void;
}

/** Footer decor of a migration cell, below focus-in: runs the migration against a seeded container. */
export const runtimeButtonDrawer: CellDrawer<RuntimeButtonData> = {
    draw(_data, cell, ctx) {
        drawButton(cell, ctx, SLOT, LABEL, COLOR);
    },

    onClick(data, cell) {
        if (!buttonHit(cell, SLOT)) return;
        data.run(data.migrationId);
        return true;
    },
};
