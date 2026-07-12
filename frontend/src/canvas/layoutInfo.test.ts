import { describe, it, expect } from 'vitest'
import type { Migration, Operation } from '../api/migrationApi'
import { buildTableOperationsMap, computeLayout, TABLE_ROW_HEIGHT } from './layoutInfo'

function createTable(tableName: string): Operation {
  return { type: 'CREATE_TABLE', migrationId: '', tableName, columns: [], isPartitioned: false, partitionOf: null }
}

function addColumn(tableName: string): Operation {
  return { type: 'ADD_COLUMN', migrationId: '', tableName, column: { name: 'c', type: 'int', constraints: [] } }
}

function migration(id: string, timestamp: number, operations: Operation[]): Migration {
  return { id, version: id, timestamp, operations: operations.map(op => ({ ...op, migrationId: id })) }
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
    const rename: Operation = { type: 'RENAME_TABLE', migrationId: '', tableName: 'old', newTableName: 'new' }
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

  it('resolves the migrations of the row directly below a table', () => {
    const migrations = [
      migration('m1', 1, [createTable('users')]),
      migration('m2', 2, [createTable('orders')]),
    ]

    const layout = computeLayout(migrations)

    const below = layout.getNextTableMigrationsOrNull('users')
    expect(below?.[0].operation.tableName).toBe('orders')
    expect(layout.getNextTableMigrationsOrNull('orders')).toBeNull()
  })
})
