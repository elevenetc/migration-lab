import { describe, it, expect } from 'vitest'
import type { MigrationTimelineResponse, PerformanceClass, RetryVerdict, RunMigrationsResult, RuntimeAnalysisResult, RuntimeVerdict, Warning } from '../api/migrationApi'
import migrationFixture from '../../../api-contracts/fixtures/migration-response.json' with { type: 'json' }
import runMigrationsFixture from '../../../api-contracts/fixtures/run-migrations-result.json' with { type: 'json' }
import runtimeFixture from '../../../api-contracts/fixtures/runtime-result.json' with { type: 'json' }

// Compile-time guard: fixtures generated from the Go models must be assignable
// to the TypeScript types. `tsc` fails here if backend and frontend types drift.
const migrationResponse: MigrationTimelineResponse = migrationFixture as MigrationTimelineResponse
const runMigrationsResult: RunMigrationsResult = runMigrationsFixture as RunMigrationsResult
const runtimeResult: RuntimeAnalysisResult = runtimeFixture as RuntimeAnalysisResult

function assertExhaustive(val: never): never {
  throw new Error(`Unhandled type: ${JSON.stringify(val)}`)
}

function validateOperationTypes(response: MigrationTimelineResponse): void {
  for (const migration of response.timeline) {
    for (const op of migration.statements.flatMap(statement => statement.operations)) {
      switch (op.type) {
        case 'CREATE_TABLE':
          break
        case 'ADD_COLUMN':
          break
        case 'ALTER_COLUMN_TYPE':
          break
        case 'SET_NOT_NULL':
          break
        case 'DROP_NOT_NULL':
          break
        case 'SET_DEFAULT':
          break
        case 'DROP_DEFAULT':
          break
        case 'RENAME_TABLE':
          break
        case 'RENAME_COLUMN':
          break
        case 'ADD_CONSTRAINT':
          break
        case 'DROP_CONSTRAINT':
          break
        case 'DROP_TABLE':
          break
        case 'DROP_COLUMN':
          break
        default:
          assertExhaustive(op)
      }
    }
  }
}

const PERFORMANCE_CLASSES: PerformanceClass[] = ['METADATA_ONLY', 'DATA_SCANNING', 'TABLE_REWRITE']

function validatePerformanceClasses(response: MigrationTimelineResponse): void {
  for (const migration of response.timeline) {
    for (const statement of migration.statements) {
      if (!PERFORMANCE_CLASSES.includes(statement.performanceClass)) {
        throw new Error(`Statement ${migration.id}#${statement.index} has unknown performanceClass: ${statement.performanceClass}`)
      }
      for (const op of statement.operations) {
        if (!PERFORMANCE_CLASSES.includes(op.performanceClass)) {
          throw new Error(`Operation ${op.type} has unknown performanceClass: ${op.performanceClass}`)
        }
      }
    }
  }
}

function validateMap(response: MigrationTimelineResponse): void {
  for (const [id, migration] of Object.entries(response.map)) {
    if (migration.id !== id) {
      throw new Error(`migrationsById key "${id}" does not match migration.id "${migration.id}"`)
    }
  }
}

function validateTimestamps(response: MigrationTimelineResponse): void {
  for (const migration of response.timeline) {
    if (typeof migration.timestamp !== 'number') {
      throw new Error(`Migration "${migration.id}" has invalid timestamp: ${migration.timestamp}`)
    }
  }
}

function validateCreateTableMap(response: MigrationTimelineResponse): void {
  for (const [tableName, entry] of Object.entries(response.createTableMap)) {
    if (entry.tableName !== tableName) {
      throw new Error(`createTableMap key "${tableName}" does not match entry.tableName "${entry.tableName}"`)
    }
  }
}

function validateWarning(warning: Warning): void {
  if (warning.type === 'ACCESS_EXCLUSIVE_LOCK') {
    if (!warning.operationId.migrationId
      || typeof warning.operationId.statementIndex !== 'number'
      || typeof warning.operationId.opIndex !== 'number') {
      throw new Error('AccessExclusiveLock missing operationId fields')
    }
    if (!warning.tableName) {
      throw new Error('AccessExclusiveLock missing tableName')
    }
    if (!warning.message) {
      throw new Error('AccessExclusiveLock missing message')
    }
  } else {
    assertExhaustive(warning.type as never)
  }
}

function validateRunMigrationsResult(result: RunMigrationsResult): void {
  if (typeof result.success !== 'boolean') {
    throw new Error(`RunMigrationsResult.success must be boolean, got: ${typeof result.success}`)
  }
  if (typeof result.message !== 'string') {
    throw new Error(`RunMigrationsResult.message must be string, got: ${typeof result.message}`)
  }
  if (typeof result.migrationsApplied !== 'number') {
    throw new Error(`RunMigrationsResult.migrationsApplied must be number, got: ${typeof result.migrationsApplied}`)
  }
}

describe('API contract: migration response fixture', () => {
  it('is non-empty', () => {
    expect(migrationResponse.timeline.length).toBeGreaterThan(0)
  })

  it('only contains known operation types', () => {
    expect(() => validateOperationTypes(migrationResponse)).not.toThrow()
  })

  it('classifies every statement and operation', () => {
    expect(() => validatePerformanceClasses(migrationResponse)).not.toThrow()
  })

  it('covers all three performance classes', () => {
    const classes = new Set(migrationResponse.timeline
      .flatMap(migration => migration.statements)
      .map(statement => statement.performanceClass))
    expect([...classes].sort()).toEqual(['DATA_SCANNING', 'METADATA_ONLY', 'TABLE_REWRITE'])
  })

  it('keys map entries by migration id', () => {
    expect(() => validateMap(migrationResponse)).not.toThrow()
  })

  it('keys createTableMap entries by table name', () => {
    expect(() => validateCreateTableMap(migrationResponse)).not.toThrow()
  })

  it('has numeric timestamps', () => {
    expect(() => validateTimestamps(migrationResponse)).not.toThrow()
  })

  it('has well-formed analysis warnings', () => {
    for (const warning of migrationResponse.analysis.warnings) {
      expect(() => validateWarning(warning)).not.toThrow()
    }
  })
})

describe('API contract: run-migrations result fixture', () => {
  it('has the expected field types', () => {
    expect(() => validateRunMigrationsResult(runMigrationsResult)).not.toThrow()
  })
})

const RUNTIME_VERDICTS: RuntimeVerdict[] = ['COMPLETED', 'EXCEEDS_DEADLINE', 'FAILED']
const RETRY_VERDICTS: RetryVerdict[] = ['SAFE_TO_RETRY', 'NEEDS_MANUAL_CLEANUP', 'FAILURE_LOOP', 'NOT_APPLICABLE']

function validateRuntimeAnalysisResult(result: RuntimeAnalysisResult): void {
  if (!RUNTIME_VERDICTS.includes(result.verdict)) {
    throw new Error(`RuntimeAnalysisResult has unknown verdict: ${result.verdict}`)
  }
  if (!RETRY_VERDICTS.includes(result.retry)) {
    throw new Error(`RuntimeAnalysisResult has unknown retry verdict: ${result.retry}`)
  }
  for (const statement of result.statements) {
    if (!RUNTIME_VERDICTS.includes(statement.verdict)) {
      throw new Error(`Statement ${statement.statementIndex} has unknown verdict: ${statement.verdict}`)
    }
    for (const measured of [statement.observedClass, statement.predictedClass]) {
      if (measured !== undefined && !PERFORMANCE_CLASSES.includes(measured)) {
        throw new Error(`Statement ${statement.statementIndex} has unknown performance class: ${measured}`)
      }
    }
  }
  for (const finding of result.findings) {
    if (!finding.type || !finding.message || !finding.operationId.migrationId) {
      throw new Error(`RuntimeFinding is missing fields: ${JSON.stringify(finding)}`)
    }
  }
}

describe('API contract: runtime result fixture', () => {
  it('has known verdicts and well-formed findings', () => {
    expect(() => validateRuntimeAnalysisResult(runtimeResult)).not.toThrow()
  })

  // Go marshals a nil slice as null, which every consumer here would have to guard.
  it('carries arrays rather than nulls for its collections', () => {
    expect(Array.isArray(runtimeResult.seeded)).toBe(true)
    expect(Array.isArray(runtimeResult.statements)).toBe(true)
    expect(Array.isArray(runtimeResult.findings)).toBe(true)
    expect(runtimeResult.statements.every(statement => Array.isArray(statement.locks))).toBe(true)
    expect(runtimeResult.statements.every(statement => Array.isArray(statement.rewrittenRelations))).toBe(true)
  })

  // Phase 1 of the migration to runtime analysis: the prediction is carried beside the
  // measurement so a disagreement is visible rather than assumed away.
  it('scores the static prediction against the measurement', () => {
    const understated = runtimeResult.statements.find(
      statement => statement.observedClass === 'TABLE_REWRITE' && statement.predictedClass === 'METADATA_ONLY',
    )
    expect(understated).toBeDefined()
    expect(understated!.rewrittenRelations.length).toBeGreaterThan(0)
    expect(runtimeResult.findings.some(finding => finding.type === 'CLASS_UNDERSTATED')).toBe(true)
  })
})
