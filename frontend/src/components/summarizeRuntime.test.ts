import {describe, expect, it} from 'vitest'
import type {RuntimeMigrationResult, StatementMeasurement} from '../api/migrationApi'
import {summarizeRuntime} from './summarizeRuntime'

function statement(overrides: Partial<StatementMeasurement>): StatementMeasurement {
  return {
    statementIndex: 0,
    sql: '',
    durationMs: 0,
    strongestLock: '',
    locks: [],
    verdict: 'COMPLETED',
    rewrittenRelations: [],
    tuplesRead: 0,
    ...overrides,
  }
}

function result(statements: StatementMeasurement[]): RuntimeMigrationResult {
  return {
    migrationId: 'V1',
    version: '1',
    deadlineMs: 5000,
    seeded: [],
    statements,
    verdict: 'COMPLETED',
    retry: 'NOT_APPLICABLE',
    findings: [],
    message: '',
  }
}

describe('summarizeRuntime', () => {
  it('adds durations up and keeps the strongest lock and the worst class', () => {
    const summary = summarizeRuntime(result([
      statement({durationMs: 10, strongestLock: 'ShareUpdateExclusiveLock', observedClass: 'DATA_SCANNING'}),
      statement({durationMs: 600, strongestLock: 'AccessExclusiveLock', observedClass: 'TABLE_REWRITE'}),
      statement({durationMs: 5, strongestLock: 'ShareLock', observedClass: 'METADATA_ONLY'}),
    ]))

    expect(summary).toEqual({durationMs: 615, strongestLock: 'AccessExclusiveLock', worstClass: 'TABLE_REWRITE'})
  })

  it('leaves lock and class empty when nothing was observed', () => {
    expect(summarizeRuntime(result([statement({durationMs: 3})]))).toEqual({
      durationMs: 3,
      strongestLock: null,
      worstClass: null,
    })
  })
})
