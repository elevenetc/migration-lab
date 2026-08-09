import {type CellCoord, focusKeyboard, type KeyboardController, selectAndShow} from 'active-grid';

const STEPS: Record<string, number> = {ArrowLeft: -1, ArrowRight: 1};
// Rows are tables, and a migration may fill several of them, so the cells themselves are the track.
const IGNORED = new Set(['ArrowUp', 'ArrowDown']);

/** Left and right walk the migration cells, column by column and row by row within a column. */
export function migrationsKeyboard(cells: readonly CellCoord[]): KeyboardController {
    return (event, grid) => {
        if (IGNORED.has(event.key)) return true;
        const step = STEPS[event.key];
        if (step === undefined) return focusKeyboard(event, grid);
        const current = grid.getFocused() ?? grid.getSelected();
        const index = current ? cells.findIndex(c => c.row === current.row && c.col === current.col) : -1;
        const next = index < 0
            ? cells[step > 0 ? 0 : cells.length - 1]
            : cells[Math.min(Math.max(index + step, 0), cells.length - 1)];
        if (next) selectAndShow(grid, next);
        return true;
    };
}
