import { ApiError } from './apiError'

declare global {
  interface Window {
    __MIGRATION_DATA__?: MigrationTimelineResponse
  }
}

export interface Column {
  name: string
  type: string
  constraints: string[]
  // Deparsed DEFAULT expression; absent when the column has no default or the
  // backend could not deparse it. `constraints` carries 'DEFAULT' either way.
  defaultExpr?: string
}

/** How expensive an operation is relative to table size; ordinal, cheapest first. */
export type PerformanceClass = 'METADATA_ONLY' | 'DATA_SCANNING' | 'TABLE_REWRITE'

/** Carried by every operation: the backend derives it from the parsed statement. */
export interface OperationBase {
  performanceClass: PerformanceClass
}

export interface CreateTable extends OperationBase {
  type: 'CREATE_TABLE'
  tableName: string
  columns: Column[]
  isPartitioned: boolean
  partitionOf: string | null
}

export interface AddColumn extends OperationBase {
  type: 'ADD_COLUMN'
  tableName: string
  column: Column
}

export interface AlterColumnType extends OperationBase {
  type: 'ALTER_COLUMN_TYPE'
  tableName: string
  columnName: string
  newType: string
}

export interface SetNotNull extends OperationBase {
  type: 'SET_NOT_NULL'
  tableName: string
  columnName: string
}

export interface DropNotNull extends OperationBase {
  type: 'DROP_NOT_NULL'
  tableName: string
  columnName: string
}

export interface SetDefault extends OperationBase {
  type: 'SET_DEFAULT'
  tableName: string
  columnName: string
  defaultValue: string
}

export interface DropDefault extends OperationBase {
  type: 'DROP_DEFAULT'
  tableName: string
  columnName: string
}

export interface RenameTable extends OperationBase {
  type: 'RENAME_TABLE'
  tableName: string
  newTableName: string
}

export interface RenameColumn extends OperationBase {
  type: 'RENAME_COLUMN'
  tableName: string
  columnName: string
  newColumnName: string
}

export interface AddConstraint extends OperationBase {
  type: 'ADD_CONSTRAINT'
  tableName: string
  constraintName: string
  constraintType: string
  // NOT VALID: existing rows are not checked, so no scan of the table.
  notValid: boolean
}

export interface DropConstraint extends OperationBase {
  type: 'DROP_CONSTRAINT'
  tableName: string
  constraintName: string
}

export interface DropTable extends OperationBase {
  type: 'DROP_TABLE'
  tableName: string
}

export interface DropColumn extends OperationBase {
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
  /** Worst performance class among the statement's operations. */
  performanceClass: PerformanceClass
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

export interface StaticAnalysisResult {
  warnings: Warning[]
}

export interface MigrationTimelineResponse {
  timeline: Migration[]
  map: Record<string, Migration>
  createTableMap: Record<string, CreateTableMapEntry>
  analysis: StaticAnalysisResult
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
    throw new ApiError('Failed to fetch migrations', response.status, response.headers.get('X-Request-ID'))
  }
  return response.json()
}

export interface DatasetsResponse {
  datasets: string[]
}

export async function fetchDatasets(): Promise<string[]> {
  const response = await fetch('/api/datasets')
  if (!response.ok) {
    throw new ApiError('Failed to fetch datasets', response.status, response.headers.get('X-Request-ID'))
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
    throw new ApiError('Failed to run migrations', response.status, response.headers.get('X-Request-ID'))
  }
  return response.json()
}

/** How the migration itself behaved when it was executed against a seeded container. */
export type RuntimeVerdict = 'COMPLETED' | 'EXCEEDS_DEADLINE' | 'FAILED'

/** What a migration cancelled at the deadline leaves behind for the next attempt. */
export type RetryVerdict = 'SAFE_TO_RETRY' | 'NEEDS_MANUAL_CLEANUP' | 'FAILURE_LOOP' | 'NOT_APPLICABLE'

export interface SeededTable {
  table: string
  rows: number
  // Set when the table could not be filled, so its measurements are of an empty table.
  error?: string
}

export interface LockObservation {
  mode: string
  relation: string
}

export interface StatementMeasurement {
  statementIndex: number
  sql: string
  durationMs: number
  strongestLock: string
  locks: LockObservation[]
  verdict: RuntimeVerdict
  error?: string
  // The class the database produced, derived from counts rather than durations. Absent when the
  // observation is not usable: the statement never completed, or the tables held no rows.
  observedClass?: PerformanceClass
  // What static analysis predicted before the run, absent for a statement it does not model.
  predictedClass?: PerformanceClass
  // Relations whose relfilenode changed, which is true if and only if they were rewritten.
  rewrittenRelations: string[]
  tuplesRead: number
}

/** Carries the same fields as a static analysis warning, raised from a measurement. */
export interface RuntimeFinding {
  type: string
  operationId: OperationId
  tableName: string
  message: string
}

export interface RuntimeAnalysisResult {
  migrationId: string
  version: string
  deadlineMs: number
  seeded: SeededTable[]
  statements: StatementMeasurement[]
  verdict: RuntimeVerdict
  retry: RetryVerdict
  findings: RuntimeFinding[]
  message: string
}

/** Runs one migration of the loaded timeline against a seeded container. */
export async function runRuntimeAnalysis(
  migrationId: string | null,
  migrationsPath: string | null,
  migration: string,
): Promise<RuntimeAnalysisResult> {
  const target = withParams('/api/migrations/runtime-analysis', { migrationId, migrationsPath, migration })
  const response = await fetch(target, { method: 'POST' })
  if (!response.ok) {
    throw new ApiError(`Runtime analysis failed: ${response.status}`, response.status, response.headers.get('X-Request-ID'))
  }
  return response.json()
}
