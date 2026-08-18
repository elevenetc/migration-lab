import { describe, it, expect } from 'vitest'
import type { Operation } from '../api/migrationApi'
import { getOperationTarget } from './getOperationTarget'
import { getOperationTitle } from './getOperationTitle'

const cases: Array<{ operation: Operation; action: string; target: string }> = [
  {
    operation: { type: 'CREATE_TABLE', tableName: 'users', columns: [], isPartitioned: false, partitionOf: null, performanceClass: 'METADATA_ONLY' },
    action: 'create table',
    target: 'users',
  },
  {
    operation: { type: 'CREATE_TABLE', tableName: 'users_2024', columns: [], isPartitioned: false, partitionOf: 'users', performanceClass: 'METADATA_ONLY' },
    action: 'add partition',
    target: 'users_2024',
  },
  {
    operation: { type: 'ADD_COLUMN', tableName: 'users', column: { name: 'email', type: 'text', constraints: [] }, performanceClass: 'METADATA_ONLY' },
    action: 'add column',
    target: 'email',
  },
  {
    operation: { type: 'ALTER_COLUMN_TYPE', tableName: 'users', columnName: 'email', newType: 'text', performanceClass: 'TABLE_REWRITE' },
    action: 'alter column type',
    target: 'email',
  },
  {
    operation: { type: 'SET_NOT_NULL', tableName: 'users', columnName: 'email', performanceClass: 'DATA_SCANNING' },
    action: 'set not-null',
    target: 'email',
  },
  {
    operation: { type: 'DROP_NOT_NULL', tableName: 'users', columnName: 'email', performanceClass: 'METADATA_ONLY' },
    action: '-not-null',
    target: 'email',
  },
  {
    operation: { type: 'SET_DEFAULT', tableName: 'users', columnName: 'email', defaultValue: "''", performanceClass: 'METADATA_ONLY' },
    action: 'set default',
    target: 'email',
  },
  {
    operation: { type: 'DROP_DEFAULT', tableName: 'users', columnName: 'email', performanceClass: 'METADATA_ONLY' },
    action: '-default',
    target: 'email',
  },
  {
    operation: { type: 'RENAME_TABLE', tableName: 'users', newTableName: 'accounts', performanceClass: 'METADATA_ONLY' },
    action: 'rename table',
    target: 'accounts',
  },
  {
    operation: { type: 'RENAME_COLUMN', tableName: 'users', columnName: 'email', newColumnName: 'mail', performanceClass: 'METADATA_ONLY' },
    action: 'rename column',
    target: 'mail',
  },
  {
    operation: { type: 'ADD_CONSTRAINT', tableName: 'users', constraintName: 'users_pk', constraintType: 'PRIMARY KEY', notValid: false, performanceClass: 'DATA_SCANNING' },
    action: 'add constraint',
    target: 'users_pk',
  },
  {
    operation: { type: 'DROP_CONSTRAINT', tableName: 'users', constraintName: 'users_pk', performanceClass: 'METADATA_ONLY' },
    action: '-constraint',
    target: 'users_pk',
  },
  {
    operation: { type: 'DROP_TABLE', tableName: 'users', performanceClass: 'METADATA_ONLY' },
    action: 'drop table',
    target: 'users',
  },
  {
    operation: { type: 'DROP_COLUMN', tableName: 'users', columnName: 'email', performanceClass: 'METADATA_ONLY' },
    action: 'drop column',
    target: 'email',
  },
]

describe('operation text', () => {
  it.each(cases)('labels $operation.type as $action of $target', ({ operation, action, target }) => {
    expect(getOperationTitle(operation)).toBe(action)
    expect(getOperationTarget(operation)).toBe(target)
  })

  it('covers every operation variant', () => {
    const covered = new Set(cases.map(({ operation }) => operation.type))
    expect([...covered].sort()).toEqual([
      'ADD_COLUMN',
      'ADD_CONSTRAINT',
      'ALTER_COLUMN_TYPE',
      'CREATE_TABLE',
      'DROP_COLUMN',
      'DROP_CONSTRAINT',
      'DROP_DEFAULT',
      'DROP_NOT_NULL',
      'DROP_TABLE',
      'RENAME_COLUMN',
      'RENAME_TABLE',
      'SET_DEFAULT',
      'SET_NOT_NULL',
    ])
  })
})
