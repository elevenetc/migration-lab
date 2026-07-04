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
}

export interface AlterTable {
  type: 'ALTER_TABLE'
  migrationId: string
  tableName: string
  addedColumns: Column[]
  droppedColumns: string[]
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

export type Operation = CreateTable | AlterTable | AlterColumnType | SetNotNull

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

export interface MigrationTimelineResponse {
  timeline: Migration[]
  map: Record<string, Migration>
  createTableMap: Record<string, CreateTableMapEntry>
}

export async function fetchMigrations(): Promise<MigrationTimelineResponse> {
  const response = await fetch('/api/migrations')
  if (!response.ok) {
    throw new Error('Failed to fetch migrations')
  }
  return response.json()
}
