export interface Column {
  name: string
  type: string
  constraints: string[]
}

export interface CreateTable {
  type: 'CREATE_TABLE'
  migrationId: string
  tableName: string
  columns: Column[]
  isPartitioned: boolean
  partitionOf: string | null
}

export interface AddColumn {
  type: 'ADD_COLUMN'
  migrationId: string
  tableName: string
  column: Column
}

export interface AlterColumnType {
  type: 'ALTER_COLUMN_TYPE'
  migrationId: string
  tableName: string
  columnName: string
  newType: string
}

export interface SetNotNull {
  type: 'SET_NOT_NULL'
  migrationId: string
  tableName: string
  columnName: string
}

export interface DropNotNull {
  type: 'DROP_NOT_NULL'
  migrationId: string
  tableName: string
  columnName: string
}

export interface SetDefault {
  type: 'SET_DEFAULT'
  migrationId: string
  tableName: string
  columnName: string
  defaultValue: string
}

export interface DropDefault {
  type: 'DROP_DEFAULT'
  migrationId: string
  tableName: string
  columnName: string
}

export interface RenameTable {
  type: 'RENAME_TABLE'
  migrationId: string
  tableName: string
  newTableName: string
}

export interface RenameColumn {
  type: 'RENAME_COLUMN'
  migrationId: string
  tableName: string
  columnName: string
  newColumnName: string
}

export interface AddConstraint {
  type: 'ADD_CONSTRAINT'
  migrationId: string
  tableName: string
  constraintName: string
  constraintType: string
}

export interface DropConstraint {
  type: 'DROP_CONSTRAINT'
  migrationId: string
  tableName: string
  constraintName: string
}

export interface DropTable {
  type: 'DROP_TABLE'
  migrationId: string
  tableName: string
}

export interface DropColumn {
  type: 'DROP_COLUMN'
  migrationId: string
  tableName: string
  columnName: string
}

export type Operation = CreateTable | AddColumn | AlterColumnType | SetNotNull | DropNotNull | SetDefault | DropDefault | RenameTable | RenameColumn | AddConstraint | DropConstraint | DropTable | DropColumn

export interface Migration {
  id: string
  version: string
  timestamp: number
  operations: Operation[]
}

export interface CreateTableMapEntry {
  migrationId: string
  tableName: string
  columns: Column[]
}

export interface OperationId {
  migrationId: string
  tableName: string
}

export interface AccessExclusiveLock {
  type: 'ACCESS_EXCLUSIVE_LOCK'
  operationId: OperationId
  tableName: string
  message: string
}

export type Warning = AccessExclusiveLock

export interface AnalysisResult {
  warnings: Warning[]
}

export interface MigrationTimelineResponse {
  timeline: Migration[]
  map: Record<string, Migration>
  createTableMap: Record<string, CreateTableMapEntry>
  analysis: AnalysisResult
}

export async function fetchMigrations(): Promise<MigrationTimelineResponse> {
  const response = await fetch('/api/migrations')
  if (!response.ok) {
    throw new Error('Failed to fetch migrations')
  }
  return response.json()
}
