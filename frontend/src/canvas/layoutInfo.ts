import type {Migration, Operation, Warning} from '../api/migrationApi'

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
        migration.operations.forEach(operation => {
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
    operation: Operation
    warnings: Warning[]
}

export interface TableLayoutInfo {
    tableName: string
    y: number
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
    // operations on a table, but they collapse into a single rectangle.
    interface Placed {
        timestampRank: number
        title: string
        migration: Migration,
        operation: Operation
        warnings: Warning[]
    }

    // Placed entries grouped per table, in row order
    const placedByTable = tableOrder.map(tableName => {
        const ops = tableOperations.get(tableName) ?? []
        const seenMigrations = new Set<string>()
        const placed: Placed[] = []
        ops.forEach(({operation, migration}) => {
            if (seenMigrations.has(migration.id)) return
            seenMigrations.add(migration.id)
            placed.push({
                timestampRank: timestampRank.get(migration.timestamp) ?? 0,
                title: migration.operations.map(op => op.type).join(', '),
                migration: migration,
                operation: operation,
                warnings: warnings.filter(w =>
                    w.operationId.migrationId === migration.id && w.operationId.tableName === tableName)
            })
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
        tableMigrations.set(table, placed.map(({timestampRank, title, migration, operation, warnings}) => ({
            x: columnX.get(timestampRank)!,
            y: table.y,
            w: rectWidth(title),
            h: TABLE_ROW_HEIGHT,
            title,
            migration,
            operation,
            warnings
        })))
    })

    const width = x
    const height = tableOrder.length * (TABLE_ROW_HEIGHT + Y_GAP)
    const tables = [...tableMigrations.keys()].map(table => table.tableName)

    return new LayoutInfo(tableMigrations, tables, width, height)
}
