import { describe, it, expect } from 'vitest'
import type { MigrationTimelineResponse, RunMigrationsResult, Warning } from '../api/migrationApi'
import migrationFixture from '../../../api-contracts/fixtures/migration-response.json' with { type: 'json' }
import runMigrationsFixture from '../../../api-contracts/fixtures/run-migrations-result.json' with { type: 'json' }

// Compile-time guard: fixtures generated from the Go models must be assignable
// to the TypeScript types. `tsc` fails here if backend and frontend types drift.
const migrationResponse: MigrationTimelineResponse = migrationFixture as MigrationTimelineResponse
const runMigrationsResult: RunMigrationsResult = runMigrationsFixture as RunMigrationsResult

function assertExhaustive(val: never): never {
  throw new Error(`Unhandled type: ${JSON.stringify(val)}`)
}

function validateOperationTypes(response: MigrationTimelineResponse): void {
  for (const migration of response.timeline) {
    for (const op of migration.operations) {
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
    if (!entry.migrationId) {
      throw new Error(`createTableMap entry for "${tableName}" missing migrationId`)
    }
  }
}

function validateWarning(warning: Warning): void {
  if (warning.type === 'ACCESS_EXCLUSIVE_LOCK') {
    if (!warning.operationId.migrationId || !warning.operationId.tableName) {
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
