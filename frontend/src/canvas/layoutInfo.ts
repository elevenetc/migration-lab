import type {Migration, Operation, Warning} from '../api/migrationApi'
import {getOperationTitle} from './getOperationTitle.ts'

export interface OperationEntry {
    migration: Migration
    operation: Operation
}

export interface TableOperations {
    tableOperations: Map<string, OperationEntry[]>
    tableOrder: string[]
}

// Groups operations by table row; rows follow first-appearance order.
// For RENAME_TABLE the new table name keys the row.
export function buildTableOperationsMap(migrations: Migration[]): TableOperations {
    const tableOperations = new Map<string, OperationEntry[]>()
    const tableOrder: string[] = []

    migrations.forEach(migration => {
        migration.statements.flatMap(statement => statement.operations).forEach(operation => {
            const tableName = operation.type === 'RENAME_TABLE'
                ? operation.newTableName
                : operation.tableName

            if (!tableOperations.has(tableName)) {
                tableOperations.set(tableName, [])
                tableOrder.push(tableName)
            }
            tableOperations.get(tableName)!.push({migration, operation})
        })
    })

    return {tableOperations, tableOrder}
}

export interface OperationLayoutInfo {
    x: number
    y: number
    w: number
    h: number
    title: string
    migration: Migration,
    // All operations of the migration on this table row, in statement order.
    operations: Operation[]
    warnings: Warning[]
}

export interface TableLayoutInfo {
    tableName: string
    y: number
    h: number
}

// Bounding box enclosing every operation rectangle of one migration.
export interface MigrationLayoutInfo {
    migration: Migration
    x: number
    y: number
    w: number
    h: number
}

// Each table row maps to its migrations in chronological order,
// so rendering code can reference sibling (next/previous) migrations.
// Iteration order of the map follows row order (top to bottom).
export class LayoutInfo {
    constructor(
        readonly tableMigrations: Map<TableLayoutInfo, OperationLayoutInfo[]>,
        // Table names in render order (top to bottom), matching map iteration order.
        readonly tables: string[],
        readonly width: number,
        readonly height: number,
        readonly migrationBounds: MigrationLayoutInfo[],
    ) {}

    // Migrations of the table rendered directly below `tableName`,
    // or null if it is the last row or not found.
    getNextTableMigrationsOrNull(tableName: string): OperationLayoutInfo[] | null {
        const index = this.tables.indexOf(tableName)
        if (index === -1 || index + 1 >= this.tables.length) return null
        const nextTableName = this.tables[index + 1]
        for (const [table, migrations] of this.tableMigrations) {
            if (table.tableName === nextTableName) return migrations
        }
        return null
    }
}

const GUTTER = 160 // left space for table-name row labels
export const TABLE_ROW_HEIGHT = 45
const Y_GAP = 3
const X_GAP = 0
const CHAR_WIDTH = 8 // ~8px per char at 13px font
const PADDING = 10

export const TRANSITION_TAG_SHIFT = 19 // height transition shift to avoid overlapping with migration tag

function rectWidth(title: string): number {
    return title.length * CHAR_WIDTH + PADDING
}

// Places one rectangle per (migration, table): rows align tables vertically,
// columns align migrations left-to-right by timestamp.
export function computeLayout(migrations: Migration[], warnings: Warning[] = []): LayoutInfo {

    const {tableOperations, tableOrder} = buildTableOperationsMap(migrations)

    const uniqueTimestamps = [...new Set(migrations.map(m => m.timestamp))].sort((a, b) => a - b)
    const timestampRank = new Map(uniqueTimestamps.map((ts, idx) => [ts, idx]))

    // One entry per (migration, table); the same migration may hold several
    // operations on a table — they share a single rectangle, stacking a flag each.
    interface Placed {
        timestampRank: number
        title: string
        migration: Migration,
        operations: Operation[]
        warnings: Warning[]
    }

    // Placed entries grouped per table, in row order
    const placedByTable = tableOrder.map(tableName => {
        const ops = tableOperations.get(tableName) ?? []
        const placedByMigration = new Map<string, Placed>()
        const placed: Placed[] = []
        ops.forEach(({operation, migration}) => {
            const existing = placedByMigration.get(migration.id)
            if (existing) {
                existing.operations.push(operation)
                return
            }
            const entry: Placed = {
                timestampRank: timestampRank.get(migration.timestamp) ?? 0,
                title: '',
                migration: migration,
                operations: [operation],
                warnings: warnings.filter(w =>
                    w.operationId.migrationId === migration.id && w.tableName === tableName)
            }
            placedByMigration.set(migration.id, entry)
            placed.push(entry)
        })
        placed.forEach(entry => {
            entry.title = entry.operations.map(getOperationTitle).join(', ')
        })
        return {tableName, placed}
    })

    // Column widths: max rectangle width per timestamp rank, across all tables
    const columnWidths = new Map<number, number>()
    placedByTable.forEach(({placed}) => placed.forEach(({timestampRank, title}) => {
        columnWidths.set(timestampRank, Math.max(columnWidths.get(timestampRank) ?? 0, rectWidth(title)))
    }))

    // Cumulative column X positions, starting after the left gutter
    const columnX = new Map<number, number>()
    let x = GUTTER
    const maxCol = columnWidths.size > 0 ? Math.max(...columnWidths.keys()) : -1
    for (let col = 0; col <= maxCol; col++) {
        columnX.set(col, x)
        x += (columnWidths.get(col) ?? 0) + X_GAP
    }

    const rowY = (rowIndex: number) => rowIndex * (TABLE_ROW_HEIGHT + Y_GAP)

    const tableMigrations = new Map<TableLayoutInfo, OperationLayoutInfo[]>()
    placedByTable.forEach(({tableName, placed}, rowIndex) => {
        const table: TableLayoutInfo = {
            tableName,
            y: rowY(rowIndex),
            h: TABLE_ROW_HEIGHT,
        }
        tableMigrations.set(table, placed.map(({timestampRank, title, migration, operations, warnings}) => ({
            x: columnX.get(timestampRank)!,
            y: table.y,
            w: rectWidth(title),
            h: TABLE_ROW_HEIGHT,
            title,
            migration,
            operations,
            warnings
        })))
    })

    const width = x
    const height = tableOrder.length * (TABLE_ROW_HEIGHT + Y_GAP)
    const tables = [...tableMigrations.keys()].map(table => table.tableName)

    return new LayoutInfo(tableMigrations, tables, width, height, computeMigrationBounds(tableMigrations))
}

// One bounding box per migration, enclosing its operation rectangles across all table rows.
function computeMigrationBounds(tableMigrations: Map<TableLayoutInfo, OperationLayoutInfo[]>): MigrationLayoutInfo[] {
    const byMigration = new Map<string, MigrationLayoutInfo>()
    for (const operations of tableMigrations.values()) {
        operations.forEach(({x, y, w, h, migration}) => {
            const bounds = byMigration.get(migration.id)
            if (!bounds) {
                byMigration.set(migration.id, {migration, x, y, w, h})
                return
            }
            const right = Math.max(bounds.x + bounds.w, x + w)
            const bottom = Math.max(bounds.y + bounds.h, y + h)
            bounds.x = Math.min(bounds.x, x)
            bounds.y = Math.min(bounds.y, y)
            bounds.w = right - bounds.x
            bounds.h = bottom - bounds.y
        })
    }
    return [...byMigration.values()]
}
