import {ActiveGrid, cellHover, cellPath} from 'active-grid';
import {BASE_MIGRATION, backgroundDrawer} from './drawers/background';
import {FOCUS_BUTTON_DECOR_HEIGHT, focusButtonDrawer} from './drawers/focusButton';
import {MIGRATION_KIND_COLOR, migrationDrawer} from './drawers/migration';
import {migrationHeaderDrawer} from './drawers/migrationHeader';
import {tableHeaderDrawer} from './drawers/tableHeader';
import {migrationFocus} from './focus/migrationFocus';
import {migrationsKeyboard} from './keyboard/migrationsKeyboard';
import {HEADER_COL, HEADER_ROW, type MigrationGridModel} from './migrationCells';

// Stacked over the background: one tint per other table row, so neighbouring tables read apart.
const ROW_TINT = 'rgba(224, 251, 252, 0.05)';
const ROW_HOVER_TINT = 'rgba(224, 251, 252, 0.14)';

/** Kind of the migration the column stands for: its cells all belong to the same migration. */
function columnKind(model: MigrationGridModel, col: number) {
    return model.cells.find(cell => cell.col === col)?.kind ?? 'alter';
}

/** Rows are tables, columns are migrations, and both axes are named by a header line of cells. */
export function buildMigrationGrid(model: MigrationGridModel): ActiveGrid {
    const grid = new ActiveGrid({rows: model.tables.length + 1, cols: model.versions.length + 1, gap: 0});

    model.tables.forEach((table, index) =>
        grid.setCell(index + 1, HEADER_COL, {data: {table}, drawer: tableHeaderDrawer}));

    model.versions.forEach((version, index) =>
        grid.setCell(HEADER_ROW, index + 1,
            {data: {name: version, kind: columnKind(model, index + 1)}, drawer: migrationHeaderDrawer}));

    model.cells.forEach(cell => {
        grid.setCell(cell.row, cell.col,
            {data: {base: BASE_MIGRATION, hoverT: 0}, drawer: backgroundDrawer},
            {data: cell, drawer: migrationDrawer});
        grid.setFooter(cell.row, cell.col, FOCUS_BUTTON_DECOR_HEIGHT, {data: null, drawer: focusButtonDrawer});
    });

    grid.setBackground(
        cellHover(ROW_TINT, 2, ROW_HOVER_TINT),
        ...model.connectors.map(({from, to, kind}) => cellPath({from, to, color: MIGRATION_KIND_COLOR[kind]})),
    );

    grid.setKeyboard(migrationsKeyboard(model.cells.map(({row, col}) => ({row, col}))));
    grid.setFocusLayout(migrationFocus);

    return grid;
}
