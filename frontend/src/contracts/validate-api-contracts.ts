import type { MigrationTimelineResponse, Warning } from '../api/migrationApi'
import migrationFixture from '../../../api-contracts/fixtures/migration-response.json' with { type: 'json' }

const _validateMigrationResponse: MigrationTimelineResponse = migrationFixture as MigrationTimelineResponse

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

function validateWarningTypes(response: MigrationTimelineResponse): void {
  for (const warning of response.analysis.warnings) {
    validateWarning(warning)
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

validateOperationTypes(_validateMigrationResponse)
validateMap(_validateMigrationResponse)
validateCreateTableMap(_validateMigrationResponse)
validateTimestamps(_validateMigrationResponse)
validateWarningTypes(_validateMigrationResponse)
