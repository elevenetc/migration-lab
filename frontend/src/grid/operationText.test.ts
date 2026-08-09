import { describe, it, expect } from 'vitest'
import type { Operation } from '../api/migrationApi'
import { getOperationTarget } from './getOperationTarget'
import { getOperationTitle } from './getOperationTitle'

const cases: Array<{ operation: Operation; action: string; target: string }> = [
  {
    operation: { type: 'CREATE_TABLE', tableName: 'users', columns: [], isPartitioned: false, partitionOf: null },
    action: 'create table',
    target: 'users',
  },
  {
    operation: { type: 'CREATE_TABLE', tableName: 'users_2024', columns: [], isPartitioned: false, partitionOf: 'users' },
    action: 'add partition',
    target: 'users_2024',
  },
  {
    operation: { type: 'ADD_COLUMN', tableName: 'users', column: { name: 'email', type: 'text', constraints: [] } },
    action: 'add column',
    target: 'email',
  },
  {
    operation: { type: 'ALTER_COLUMN_TYPE', tableName: 'users', columnName: 'email', newType: 'text' },
    action: 'alter column type',
    target: 'email',
  },
  {
    operation: { type: 'SET_NOT_NULL', tableName: 'users', columnName: 'email' },
    action: 'set not-null',
    target: 'email',
  },
  {
    operation: { type: 'DROP_NOT_NULL', tableName: 'users', columnName: 'email' },
    action: '-not-null',
    target: 'email',
  },
  {
    operation: { type: 'SET_DEFAULT', tableName: 'users', columnName: 'email', defaultValue: "''" },
    action: 'set default',
    target: 'email',
  },
  {
    operation: { type: 'DROP_DEFAULT', tableName: 'users', columnName: 'email' },
    action: '-default',
    target: 'email',
  },
  {
    operation: { type: 'RENAME_TABLE', tableName: 'users', newTableName: 'accounts' },
    action: 'rename table',
    target: 'accounts',
  },
  {
    operation: { type: 'RENAME_COLUMN', tableName: 'users', columnName: 'email', newColumnName: 'mail' },
    action: 'rename column',
    target: 'mail',
  },
  {
    operation: { type: 'ADD_CONSTRAINT', tableName: 'users', constraintName: 'users_pk', constraintType: 'PRIMARY KEY' },
    action: 'add constraint',
    target: 'users_pk',
  },
  {
    operation: { type: 'DROP_CONSTRAINT', tableName: 'users', constraintName: 'users_pk' },
    action: '-constraint',
    target: 'users_pk',
  },
  {
    operation: { type: 'DROP_TABLE', tableName: 'users' },
    action: 'drop table',
    target: 'users',
  },
  {
    operation: { type: 'DROP_COLUMN', tableName: 'users', columnName: 'email' },
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
