import { describe, it, expect } from 'vitest'
import type { Migration, Operation, Statement, Warning } from '../api/migrationApi'
import { buildMigrationCells } from './migrationCells'

function createTable(tableName: string, partitionOf: string | null = null): Operation {
  return { type: 'CREATE_TABLE', tableName, columns: [], isPartitioned: false, partitionOf, performanceClass: 'METADATA_ONLY' }
}

function addColumn(tableName: string, name = 'c'): Operation {
  return { type: 'ADD_COLUMN', tableName, column: { name, type: 'int', constraints: [] }, performanceClass: 'METADATA_ONLY' }
}

function statement(index: number, sql: string, operations: Operation[]): Statement {
  return { index, kind: 'ALTER_TABLE', sql, operations, performanceClass: 'METADATA_ONLY' }
}

function migration(id: string, statements: Statement[]): Migration {
  return { id, version: id, timestamp: 1, statements }
}

function single(id: string, operations: Operation[], sql = `-- ${id}`): Migration {
  return migration(id, [statement(0, sql, operations)])
}

describe('buildMigrationCells', () => {
  it('orders table rows by first appearance', () => {
    const model = buildMigrationCells([
      single('m1', [createTable('users')]),
      single('m2', [createTable('orders')]),
      single('m3', [addColumn('users')]),
    ])

    expect(model.tables).toEqual(['users', 'orders'])
    expect(model.cells.map(({ table, row }) => [table, row])).toEqual([
      ['users', 1],
      ['orders', 2],
      ['users', 1],
    ])
  })

  it('keys a renamed table row by its new name', () => {
    const rename: Operation = { type: 'RENAME_TABLE', tableName: 'old', newTableName: 'new', performanceClass: 'METADATA_ONLY' }

    const model = buildMigrationCells([single('m1', [rename])])

    expect(model.tables).toEqual(['new'])
  })

  it('places one column per migration in timeline order', () => {
    const model = buildMigrationCells([
      single('m1', [createTable('users')]),
      single('m2', [addColumn('users')]),
    ])

    expect(model.versions).toEqual(['m1', 'm2'])
    expect(model.cells.map(({ col }) => col)).toEqual([1, 2])
  })

  it('fills one column with a cell per table a migration touches', () => {
    const model = buildMigrationCells([single('m1', [createTable('users'), createTable('orders')])])

    expect(model.cells).toHaveLength(2)
    expect(model.cells.map(({ col }) => col)).toEqual([1, 1])
    expect(model.cells.map(({ row }) => row)).toEqual([1, 2])
  })

  it('keeps every operation of a migration on one table in a single cell', () => {
    const model = buildMigrationCells([
      single('m1', [addColumn('users', 'a'), addColumn('users', 'b'), addColumn('orgs')]),
    ])

    const users = model.cells.filter(({ table }) => table === 'users')
    expect(users).toHaveLength(1)
    expect(users[0].operations).toEqual([
      { action: 'add column', target: 'a' },
      { action: 'add column', target: 'b' },
    ])
  })

  it('derives the kind of each cell from its operations', () => {
    const drop: Operation = { type: 'DROP_TABLE', tableName: 'orders', performanceClass: 'METADATA_ONLY' }
    const model = buildMigrationCells([
      single('m1', [createTable('users')]),
      single('m2', [addColumn('users')]),
      single('m3', [createTable('users_2024', 'users')]),
      single('m4', [createTable('orders')]),
      single('m5', [drop]),
    ])

    expect(model.cells.map(({ kind }) => kind)).toEqual(['create', 'alter', 'partition', 'create', 'drop'])
  })

  it('connects a partition to the parent table latest earlier cell', () => {
    const model = buildMigrationCells([
      single('m1', [createTable('users')]),
      single('m2', [addColumn('users')]),
      single('m3', [createTable('users_2024', 'users')]),
    ])

    expect(model.connectors).toEqual([
      { from: { row: 1, col: 2 }, to: { row: 2, col: 3 }, kind: 'partition' },
    ])
  })

  it('connects a renamed table to the old table row', () => {
    const rename: Operation = { type: 'RENAME_TABLE', tableName: 'old', newTableName: 'new', performanceClass: 'METADATA_ONLY' }
    const model = buildMigrationCells([single('m1', [createTable('old')]), single('m2', [rename])])

    expect(model.connectors).toEqual([
      { from: { row: 1, col: 1 }, to: { row: 2, col: 2 }, kind: 'alter' },
    ])
  })

  it('omits a connector when the source row has no earlier cell', () => {
    const model = buildMigrationCells([single('m1', [createTable('users_2024', 'users')])])

    expect(model.connectors).toEqual([])
  })

  it('attaches a warning to the cell of its migration and table', () => {
    const warning: Warning = {
      type: 'ACCESS_EXCLUSIVE_LOCK',
      operationId: { migrationId: 'm2', statementIndex: 0, opIndex: -1 },
      tableName: 'users',
      message: 'locks the table',
    }
    const model = buildMigrationCells([
      single('m1', [createTable('users')]),
      single('m2', [addColumn('users'), createTable('orders')]),
    ], [warning])

    const flagged = model.cells.filter(({ warnings }) => warnings.length > 0)
    expect(flagged).toHaveLength(1)
    expect(flagged[0]).toMatchObject({ version: 'm2', table: 'users' })
    expect(flagged[0].warnings).toEqual([{ title: 'exclusive lock', message: 'locks the table' }])
  })

  it('warns on data-scanning and table-rewrite operations, after the backend warnings', () => {
    const warning: Warning = {
      type: 'ACCESS_EXCLUSIVE_LOCK',
      operationId: { migrationId: 'm1', statementIndex: 0, opIndex: -1 },
      tableName: 'users',
      message: 'locks the table',
    }
    const setNotNull: Operation = { type: 'SET_NOT_NULL', tableName: 'users', columnName: 'email', performanceClass: 'DATA_SCANNING' }
    const alterType: Operation = { type: 'ALTER_COLUMN_TYPE', tableName: 'users', columnName: 'email', newType: 'text', performanceClass: 'TABLE_REWRITE' }

    const model = buildMigrationCells([single('m1', [addColumn('users'), setNotNull, alterType])], [warning])

    expect(model.cells[0].warnings.map(({ title }) => title))
      .toEqual(['exclusive lock', 'data scan', 'table rewrite'])
  })

  it('raises no warning for metadata-only operations', () => {
    const model = buildMigrationCells([single('m1', [createTable('users'), addColumn('users')])])

    expect(model.cells[0].warnings).toEqual([])
  })

  it('collects the sql of every statement touching the table, once per statement', () => {
    const model = buildMigrationCells([
      migration('m1', [
        statement(0, 'CREATE TABLE users ();', [createTable('users')]),
        statement(1, 'ALTER TABLE users ADD a, ADD b;', [addColumn('users', 'a'), addColumn('users', 'b')]),
        statement(2, 'CREATE TABLE orders ();', [createTable('orders')]),
      ]),
    ])

    const [users, orders] = model.cells
    expect(users.sql).toEqual(['CREATE TABLE users ();', 'ALTER TABLE users ADD a, ADD b;'])
    expect(orders.sql).toEqual(['CREATE TABLE orders ();'])
  })

  it('returns an empty model for no migrations', () => {
    expect(buildMigrationCells([])).toEqual({ tables: [], versions: [], cells: [], connectors: [] })
  })
})
