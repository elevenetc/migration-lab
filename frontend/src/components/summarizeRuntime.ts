import type {PerformanceClass, RuntimeMigrationResult} from '../api/migrationApi'

// PostgreSQL table lock modes, weakest first.
const LOCK_ORDER = [
  'AccessShareLock',
  'RowShareLock',
  'RowExclusiveLock',
  'ShareUpdateExclusiveLock',
  'ShareLock',
  'ShareRowExclusiveLock',
  'ExclusiveLock',
  'AccessExclusiveLock',
]

const CLASS_ORDER: PerformanceClass[] = ['METADATA_ONLY', 'DATA_SCANNING', 'TABLE_REWRITE']

export interface RuntimeSummary {
  durationMs: number
  strongestLock: string | null
  worstClass: PerformanceClass | null
}

function strongest<T>(values: T[], order: T[]): T | null {
  return values.reduce<T | null>(
    (best, value) => (best === null || order.indexOf(value) > order.indexOf(best) ? value : best),
    null,
  )
}

/** One line of a measured timeline: what the migration cost in total, at its worst. */
export function summarizeRuntime(result: RuntimeMigrationResult): RuntimeSummary {
  return {
    durationMs: result.statements.reduce((total, statement) => total + statement.durationMs, 0),
    strongestLock: strongest(result.statements.map(statement => statement.strongestLock).filter(Boolean), LOCK_ORDER),
    worstClass: strongest(
      result.statements.flatMap(statement => (statement.observedClass ? [statement.observedClass] : [])),
      CLASS_ORDER,
    ),
  }
}
