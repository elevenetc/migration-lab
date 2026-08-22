package models

// RuntimeVerdict is how a migration behaved when it was actually executed.
type RuntimeVerdict string

const (
	// RuntimeCompleted means every statement finished inside the deadline.
	RuntimeCompleted RuntimeVerdict = "COMPLETED"
	// RuntimeExceedsDeadline means a statement was still running when the deadline hit
	// and was cancelled, the way a pod is killed mid-migration.
	RuntimeExceedsDeadline RuntimeVerdict = "EXCEEDS_DEADLINE"
	// RuntimeFailed means a statement raised an error unrelated to the deadline.
	RuntimeFailed RuntimeVerdict = "FAILED"
)

// RetryVerdict is what a cancelled migration leaves behind for the next attempt.
type RetryVerdict string

const (
	RetryManualCleanup RetryVerdict = "NEEDS_MANUAL_CLEANUP"
	RetryFailureLoop   RetryVerdict = "FAILURE_LOOP"
	RetryNotApplicable RetryVerdict = "NOT_APPLICABLE"
)

// Runtime finding types; the shape mirrors a static analysis warning so both
// groups render through one path.
const (
	FindingExceedsDeadline  = "EXCEEDS_DEADLINE"
	FindingBlocksReaders    = "BLOCKS_READERS"
	FindingExclusiveLock    = "EXCLUSIVE_LOCK_HELD"
	FindingInvalidIndexLeft = "INVALID_INDEX_LEFT"
	FindingSeedFailed       = "SEED_FAILED"
	FindingStatementFailed  = "STATEMENT_FAILED"
)

// SeededTable is one table filled before the migration ran. Error is set when
// the table could not be seeded, so its measurements read as unloaded.
type SeededTable struct {
	Table string `json:"table"`
	Rows  int64  `json:"rows"`
	Error string `json:"error,omitempty"`
}

// LockObservation is one relation lock the migration's backend held, sampled
// from pg_locks while the statement ran.
type LockObservation struct {
	Mode     string `json:"mode"`
	Relation string `json:"relation"`
}

// ProbeResult is what a concurrent reader session saw while the migration ran.
type ProbeResult struct {
	Table        string `json:"table"`
	Samples      int    `json:"samples"`
	Errors       int    `json:"errors"`
	MaxLatencyMs int64  `json:"maxLatencyMs"`
	// How long the reader was observed waiting on a lock the migration held, read
	// from PostgreSQL rather than inferred from read latency.
	BlockedMs int64 `json:"blockedMs"`
}

// StatementMeasurement is one executed statement of the analyzed migration.
type StatementMeasurement struct {
	StatementIndex int               `json:"statementIndex"`
	SQL            string            `json:"sql"`
	DurationMs     int64             `json:"durationMs"`
	StrongestLock  string            `json:"strongestLock"`
	Locks          []LockObservation `json:"locks"`
	Verdict        RuntimeVerdict    `json:"verdict"`
	Error          string            `json:"error,omitempty"`
}

// RuntimeFinding carries the same fields as a static analysis warning, so the
// two analysis groups stay renderable through one path.
type RuntimeFinding struct {
	Type        string      `json:"type"`
	OperationID OperationID `json:"operationId"`
	TableName   string      `json:"tableName"`
	Message     string      `json:"message"`
}

// RuntimeResult is one migration executed against a seeded container.
type RuntimeResult struct {
	MigrationID string                 `json:"migrationId"`
	Version     string                 `json:"version"`
	DeadlineMs  int64                  `json:"deadlineMs"`
	Seeded      []SeededTable          `json:"seeded"`
	Statements  []StatementMeasurement `json:"statements"`
	Probes      []ProbeResult          `json:"probes"`
	Verdict     RuntimeVerdict         `json:"verdict"`
	Retry       RetryVerdict           `json:"retry"`
	Findings    []RuntimeFinding       `json:"findings"`
	Message     string                 `json:"message"`
}
