import { describe, it, expect } from 'vitest'
import type { Operation } from '../api/migrationApi'
import { getColorFromString } from './getColorFromString'
import { getOperationColor } from './getOperationColor'
import { getOperationTitle } from './getOperationTitle'
import { getTableColor } from './getTableColor'

describe('getColorFromString', () => {
  it('is deterministic for the same input', () => {
    expect(getColorFromString('users')).toBe(getColorFromString('users'))
  })

  it('returns a well-formed hsl color', () => {
    expect(getColorFromString('orders')).toMatch(/^hsl\(\d{1,3}, 65%, 55%\)$/)
  })

  it('gives different hues to different strings', () => {
    expect(getColorFromString('users')).not.toBe(getColorFromString('orders'))
  })
})

describe('getOperationColor', () => {
  it('colors additive operations green', () => {
    const create: Operation = { type: 'CREATE_TABLE', migrationId: 'm', tableName: 't', columns: [], isPartitioned: false, partitionOf: null }
    expect(getOperationColor(create)).toBe('#28b828')
  })

  it('colors destructive operations orange', () => {
    const drop: Operation = { type: 'DROP_TABLE', migrationId: 'm', tableName: 't' }
    expect(getOperationColor(drop)).toBe('#ff5900')
  })

  it('colors mutating operations blue', () => {
    const rename: Operation = { type: 'RENAME_TABLE', migrationId: 'm', tableName: 't', newTableName: 't2' }
    expect(getOperationColor(rename)).toBe('#2682cd')
  })
})

describe('getOperationTitle', () => {
  it('labels a plain create table', () => {
    const create: Operation = { type: 'CREATE_TABLE', migrationId: 'm', tableName: 't', columns: [], isPartitioned: false, partitionOf: null }
    expect(getOperationTitle(create)).toBe('create table')
  })

  it('labels a partition create table as add partition', () => {
    const partition: Operation = { type: 'CREATE_TABLE', migrationId: 'm', tableName: 't', columns: [], isPartitioned: false, partitionOf: 'parent' }
    expect(getOperationTitle(partition)).toBe('add partition')
  })

  it('labels an add column', () => {
    const add: Operation = { type: 'ADD_COLUMN', migrationId: 'm', tableName: 't', column: { name: 'c', type: 'int', constraints: [] } }
    expect(getOperationTitle(add)).toBe('add column')
  })
})

describe('getTableColor', () => {
  it('keys a renamed table by its new name', () => {
    const rename: Operation = { type: 'RENAME_TABLE', migrationId: 'm', tableName: 'old', newTableName: 'new' }
    expect(getTableColor(rename)).toBe(getColorFromString('new'))
  })
})
