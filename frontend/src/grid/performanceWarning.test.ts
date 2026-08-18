import { describe, it, expect } from 'vitest'
import type { Operation } from '../api/migrationApi'
import { getPerformanceWarning } from './getPerformanceWarning'

describe('getPerformanceWarning', () => {
  it('raises nothing for a metadata-only operation', () => {
    const operation: Operation = { type: 'DROP_COLUMN', tableName: 'users', columnName: 'email', performanceClass: 'METADATA_ONLY' }

    expect(getPerformanceWarning(operation)).toBeNull()
  })

  it('names the operation in a data-scanning warning', () => {
    const operation: Operation = { type: 'SET_NOT_NULL', tableName: 'users', columnName: 'email', performanceClass: 'DATA_SCANNING' }

    const warning = getPerformanceWarning(operation)

    expect(warning?.title).toBe('data scan')
    expect(warning?.message).toContain('set not-null email')
  })

  it('names the operation in a table-rewrite warning', () => {
    const operation: Operation = { type: 'ALTER_COLUMN_TYPE', tableName: 'users', columnName: 'email', newType: 'text', performanceClass: 'TABLE_REWRITE' }

    const warning = getPerformanceWarning(operation)

    expect(warning?.title).toBe('table rewrite')
    expect(warning?.message).toContain('alter column type email')
  })
})
