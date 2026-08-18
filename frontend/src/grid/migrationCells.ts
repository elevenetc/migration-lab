import type {CellCoord} from 'active-grid';
import type {Migration, Operation, Warning} from '../api/migrationApi.ts';
import type {CellWarning, MigrationData, MigrationKind} from './drawers/migration';
import {getMigrationKind} from './getMigrationKind';
import {getOperationTarget} from './getOperationTarget';
import {getOperationTitle} from './getOperationTitle';
import {getPerformanceWarning} from './getPerformanceWarning';
import {getWarningTitle} from './getWarningTitle';

// Row 0 names the migrations and column 0 names the tables, so both axes are offset by one.
export const HEADER_ROW = 0;
export const HEADER_COL = 0;

/** One migration applied to one table, placed on the grid. */
export interface MigrationCell extends MigrationData {
    row: number;
    col: number;
}

/** Routed line from the cell a table derives from into the cell that derives it. */
export interface Connector {
    from: CellCoord;
    to: CellCoord;
    kind: MigrationKind;
}

export interface MigrationGridModel {
    tables: string[];
    versions: string[];
    cells: MigrationCell[];
    connectors: Connector[];
}

/** Row an operation lands on; a renamed table is keyed by its new name. */
function tableOf(operation: Operation): string {
    return operation.type === 'RENAME_TABLE' ? operation.newTableName : operation.tableName;
}

/** Table the operation derives its own table from, connected by a routed line. */
function sourceOf(operation: Operation): string | null {
    if (operation.type === 'CREATE_TABLE') return operation.partitionOf;
    if (operation.type === 'RENAME_TABLE') return operation.tableName;
    return null;
}

/** Table rows in first-appearance order. */
function tableOrder(migrations: Migration[]): string[] {
    const tables: string[] = [];
    migrations.forEach(migration => {
        migration.statements.flatMap(({operations}) => operations).forEach(operation => {
            const table = tableOf(operation);
            if (!tables.includes(table)) tables.push(table);
        });
    });
    return tables;
}

interface TableGroup {
    operations: Operation[];
    /** SQL of every statement of the migration touching the table, in statement order. */
    sql: string[];
    lastStatement: number;
}

/** Operations of one migration grouped by the table row they land on, in statement order. */
function groupByTable(migration: Migration): Map<string, TableGroup> {
    const groups = new Map<string, TableGroup>();
    migration.statements.forEach(statement => {
        statement.operations.forEach(operation => {
            const table = tableOf(operation);
            const group = groups.get(table) ?? {operations: [], sql: [], lastStatement: -1};
            group.operations.push(operation);
            if (group.lastStatement !== statement.index) {
                group.sql.push(statement.sql);
                group.lastStatement = statement.index;
            }
            groups.set(table, group);
        });
    });
    return groups;
}

function warningsOf(warnings: Warning[], migrationId: string, table: string): CellWarning[] {
    return warnings
        .filter(warning => warning.operationId.migrationId === migrationId && warning.tableName === table)
        .map(warning => ({title: getWarningTitle(warning), message: warning.message}));
}

/** Backend warnings of the cell, then the performance class of each of its operations. */
function cellWarnings(warnings: Warning[], migrationId: string, table: string, operations: Operation[]): CellWarning[] {
    return [
        ...warningsOf(warnings, migrationId, table),
        ...operations.map(getPerformanceWarning).filter((warning): warning is CellWarning => warning !== null),
    ];
}

/** Column of the row's latest cell left of `col`, where a connector into `col` starts. */
function latestBefore(colsByRow: Map<number, number[]>, row: number, col: number): number | null {
    const before = (colsByRow.get(row) ?? []).filter(candidate => candidate < col);
    return before.length > 0 ? before[before.length - 1] : null;
}

/**
 * Rows are tables, columns are migrations: one cell per (migration, table) pair,
 * holding every operation the migration applies to that table.
 */
export function buildMigrationCells(migrations: Migration[], warnings: Warning[] = []): MigrationGridModel {
    const tables = tableOrder(migrations);
    const rowOf = (table: string) => tables.indexOf(table) + 1;

    const placed = migrations.flatMap((migration, index) => [...groupByTable(migration)]
        .map(([table, group]) => ({table, group, migration, row: rowOf(table), col: index + 1}))
        .sort((a, b) => a.row - b.row));

    const cells: MigrationCell[] = placed.map(({table, group, migration, row, col}) => ({
        row,
        col,
        table,
        version: migration.version,
        kind: getMigrationKind(group.operations),
        operations: group.operations.map(operation => ({
            action: getOperationTitle(operation),
            target: getOperationTarget(operation),
        })),
        sql: group.sql,
        warnings: cellWarnings(warnings, migration.id, table, group.operations),
    }));

    // Columns hold cells in ascending order, so each row's columns come out sorted.
    const colsByRow = new Map<number, number[]>();
    cells.forEach(({row, col}) => colsByRow.set(row, [...(colsByRow.get(row) ?? []), col]));

    const connectors = cells.flatMap((cell, index) => {
        const sources = placed[index].group.operations
            .map(sourceOf)
            .filter((source): source is string => source !== null && tables.includes(source));
        return sources.flatMap(source => {
            const row = rowOf(source);
            const col = row === cell.row ? null : latestBefore(colsByRow, row, cell.col);
            return col === null ? [] : [{from: {row, col}, to: {row: cell.row, col: cell.col}, kind: cell.kind}];
        });
    });

    return {tables, versions: migrations.map(({version}) => version), cells, connectors};
}
