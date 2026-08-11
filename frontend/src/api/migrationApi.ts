declare global {
  interface Window {
    __MIGRATION_DATA__?: MigrationTimelineResponse
  }
}

export interface Column {
  name: string
  type: string
  constraints: string[]
}

export interface CreateTable {
  type: 'CREATE_TABLE'
  tableName: string
  columns: Column[]
  isPartitioned: boolean
  partitionOf: string | null
}

export interface AddColumn {
  type: 'ADD_COLUMN'
  tableName: string
  column: Column
}

export interface AlterColumnType {
  type: 'ALTER_COLUMN_TYPE'
  tableName: string
  columnName: string
  newType: string
}

export interface SetNotNull {
  type: 'SET_NOT_NULL'
  tableName: string
  columnName: string
}

export interface DropNotNull {
  type: 'DROP_NOT_NULL'
  tableName: string
  columnName: string
}

export interface SetDefault {
  type: 'SET_DEFAULT'
  tableName: string
  columnName: string
  defaultValue: string
}

export interface DropDefault {
  type: 'DROP_DEFAULT'
  tableName: string
  columnName: string
}

export interface RenameTable {
  type: 'RENAME_TABLE'
  tableName: string
  newTableName: string
}

export interface RenameColumn {
  type: 'RENAME_COLUMN'
  tableName: string
  columnName: string
  newColumnName: string
}

export interface AddConstraint {
  type: 'ADD_CONSTRAINT'
  tableName: string
  constraintName: string
  constraintType: string
}

export interface DropConstraint {
  type: 'DROP_CONSTRAINT'
  tableName: string
  constraintName: string
}

export interface DropTable {
  type: 'DROP_TABLE'
  tableName: string
}

export interface DropColumn {
  type: 'DROP_COLUMN'
  tableName: string
  columnName: string
}

export type Operation = CreateTable | AddColumn | AlterColumnType | SetNotNull | DropNotNull | SetDefault | DropDefault | RenameTable | RenameColumn | AddConstraint | DropConstraint | DropTable | DropColumn

export type StatementKind = 'CREATE_TABLE' | 'ALTER_TABLE' | 'DROP_TABLE' | 'RENAME'

export interface Statement {
  index: number
  kind: StatementKind
  sql: string
  operations: Operation[]
}

export interface Migration {
  id: string
  version: string
  timestamp: number
  statements: Statement[]
}

export interface CreateTableMapEntry {
  tableName: string
  columns: Column[]
}

export interface OperationId {
  migrationId: string
  statementIndex: number
  // -1 for a statement-scoped warning
  opIndex: number
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

function withParams(path: string, params: Record<string, string | null>): string {
  const query = new URLSearchParams()
  for (const [key, value] of Object.entries(params)) {
    if (value) query.set(key, value)
  }
  const search = query.toString()
  return search ? `${path}?${search}` : path
}

export async function fetchMigrations(migrationId: string | null, migrationsPath: string | null): Promise<MigrationTimelineResponse> {
  if (window.__MIGRATION_DATA__) {
    return window.__MIGRATION_DATA__
  }
  const response = await fetch(withParams('/api/migrations', { migrationId, migrationsPath }))
  if (!response.ok) {
    throw new Error('Failed to fetch migrations')
  }
  return response.json()
}

export interface DatasetsResponse {
  datasets: string[]
}

export async function fetchDatasets(): Promise<string[]> {
  const response = await fetch('/api/datasets')
  if (!response.ok) {
    throw new Error('Failed to fetch datasets')
  }
  const body: DatasetsResponse = await response.json()
  return body.datasets
}

export interface RunMigrationsResult {
  success: boolean
  message: string
  migrationsApplied: number
}

export async function runMigrations(migrationId: string | null): Promise<RunMigrationsResult> {
  const response = await fetch(withParams('/api/migrations/run', { migrationId }), { method: 'POST' })
  if (!response.ok) {
    throw new Error('Failed to run migrations')
  }
  return response.json()
}
