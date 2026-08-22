import {type CellDrawer} from 'active-grid';
import {buttonHit, drawButton} from './cellButton';

const SLOT = 0;
const LABEL = 'focus-in';
const COLOR = '#98c1d9';

/** Footer decor of a migration cell: zooms the grid into the cell. */
export const focusButtonDrawer: CellDrawer<null> = {
    draw(_data, cell, ctx) {
        drawButton(cell, ctx, SLOT, LABEL, COLOR);
    },

    onClick(_data, cell) {
        if (!buttonHit(cell, SLOT)) return;
        cell.grid.focusCell(cell.row, cell.col);
        return true;
    },
};
