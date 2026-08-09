import { describe, it, expect } from 'vitest'
import type { Migration, Operation, Warning } from '../api/migrationApi'
import { buildTableOperationsMap, computeLayout, TABLE_ROW_HEIGHT } from './layoutInfo'

function createTable(tableName: string): Operation {
  return { type: 'CREATE_TABLE', tableName, columns: [], isPartitioned: false, partitionOf: null }
}

function addColumn(tableName: string): Operation {
  return { type: 'ADD_COLUMN', tableName, column: { name: 'c', type: 'int', constraints: [] } }
}

function migration(id: string, timestamp: number, operations: Operation[]): Migration {
  return { id, version: id, timestamp, statements: [{ index: 0, kind: 'ALTER_TABLE', sql: '', operations }] }
}

describe('buildTableOperationsMap', () => {
  it('groups operations per table in first-appearance order', () => {
    const migrations = [
      migration('m1', 1, [createTable('users')]),
      migration('m2', 2, [createTable('orders')]),
      migration('m3', 3, [addColumn('users')]),
    ]

    const { tableOrder, tableOperations } = buildTableOperationsMap(migrations)

    expect(tableOrder).toEqual(['users', 'orders'])
    expect(tableOperations.get('users')).toHaveLength(2)
    expect(tableOperations.get('orders')).toHaveLength(1)
  })

  it('keys a renamed table row by its new name', () => {
    const rename: Operation = { type: 'RENAME_TABLE', tableName: 'old', newTableName: 'new' }
    const migrations = [migration('m1', 1, [rename])]

    const { tableOrder } = buildTableOperationsMap(migrations)

    expect(tableOrder).toEqual(['new'])
  })
})

describe('computeLayout', () => {
  it('aligns migrations with the same timestamp in the same column', () => {
    const migrations = [
      migration('m1', 100, [createTable('users')]),
      migration('m2', 100, [createTable('orders')]),
    ]

    const layout = computeLayout(migrations)
    const rects = [...layout.tableMigrations.values()].flat()

    expect(rects).toHaveLength(2)
    const [users, orders] = rects
    expect(users.x).toBe(orders.x) // same timestamp -> same column
    expect(users.y).not.toBe(orders.y) // different tables -> different rows
  })

  it('places later timestamps in columns further right', () => {
    const migrations = [
      migration('m1', 100, [createTable('users')]),
      migration('m2', 200, [addColumn('users')]),
    ]

    const rects = [...computeLayout(migrations).tableMigrations.values()].flat()

    expect(rects[1].x).toBeGreaterThan(rects[0].x)
  })

  it('exposes tables in top-to-bottom render order', () => {
    const migrations = [
      migration('m1', 1, [createTable('users')]),
      migration('m2', 2, [createTable('orders')]),
    ]

    const layout = computeLayout(migrations)

    expect(layout.tables).toEqual(['users', 'orders'])
    expect(layout.height).toBe(2 * (TABLE_ROW_HEIGHT + 3))
  })

  it('attaches a warning to the rectangle of its migration and table', () => {
    const migrations = [
      migration('m1', 1, [createTable('users')]),
      migration('m2', 2, [addColumn('users'), createTable('orders')]),
    ]
    const warning: Warning = {
      type: 'ACCESS_EXCLUSIVE_LOCK',
      operationId: { migrationId: 'm2', statementIndex: 0, opIndex: -1 },
      tableName: 'users',
      message: 'lock',
    }

    const rects = [...computeLayout(migrations, [warning]).tableMigrations.values()].flat()

    const flagged = rects.filter(r => r.warnings.length > 0)
    expect(flagged).toHaveLength(1)
    expect(flagged[0].migration.id).toBe('m2')
    expect(flagged[0].operations[0].tableName).toBe('users')
  })

  it('keeps every operation of a migration on the same table in one rectangle', () => {
    const migrations = [
      migration('m1', 1, [createTable('users')]),
      migration('m2', 2, [addColumn('users'), addColumn('users'), addColumn('orgs')]),
    ]

    const rects = [...computeLayout(migrations).tableMigrations.values()].flat()

    const usersM2 = rects.filter(r => r.migration.id === 'm2' && r.operations[0].tableName === 'users')
    expect(usersM2).toHaveLength(1)
    expect(usersM2[0].operations.map(o => o.type)).toEqual(['ADD_COLUMN', 'ADD_COLUMN'])
  })

  it('bounds a migration around its operation rectangles across tables', () => {
    const migrations = [
      migration('m1', 1, [createTable('users'), createTable('orders')]),
      migration('m2', 2, [addColumn('users')]),
    ]

    const layout = computeLayout(migrations)
    const rects = [...layout.tableMigrations.values()].flat()
    const m1Rects = rects.filter(r => r.migration.id === 'm1')

    expect(layout.migrationBounds).toHaveLength(2)
    const m1 = layout.migrationBounds.find(b => b.migration.id === 'm1')!
    expect(m1.x).toBe(Math.min(...m1Rects.map(r => r.x)))
    expect(m1.y).toBe(Math.min(...m1Rects.map(r => r.y)))
    expect(m1.x + m1.w).toBe(Math.max(...m1Rects.map(r => r.x + r.w)))
    expect(m1.y + m1.h).toBe(Math.max(...m1Rects.map(r => r.y + r.h)))
    expect(m1.h).toBeGreaterThan(TABLE_ROW_HEIGHT) // spans two rows
  })

  it('resolves the migrations of the row directly below a table', () => {
    const migrations = [
      migration('m1', 1, [createTable('users')]),
      migration('m2', 2, [createTable('orders')]),
    ]

    const layout = computeLayout(migrations)

    const below = layout.getNextTableMigrationsOrNull('users')
    expect(below?.[0].operations[0].tableName).toBe('orders')
    expect(layout.getNextTableMigrationsOrNull('orders')).toBeNull()
  })
})
