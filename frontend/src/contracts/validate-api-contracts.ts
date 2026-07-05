import type { MigrationTimelineResponse } from '../api/migrationApi'
import migrationFixture from '../../../api-contracts/fixtures/migration-response.json' with { type: 'json' }

const _validateMigrationResponse: MigrationTimelineResponse = migrationFixture as MigrationTimelineResponse

function assertExhaustive(op: never): never {
  throw new Error(`Unhandled operation type: ${JSON.stringify(op)}`)
}

function validateOperationTypes(response: MigrationTimelineResponse): void {
  for (const migration of response.timeline) {
    for (const op of migration.operations) {
      switch (op.type) {
        case 'CREATE_TABLE':
          break
        case 'ALTER_TABLE':
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

validateOperationTypes(_validateMigrationResponse)
validateMap(_validateMigrationResponse)
validateCreateTableMap(_validateMigrationResponse)
validateTimestamps(_validateMigrationResponse)
