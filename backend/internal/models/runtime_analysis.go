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
	FindingExclusiveLock    = "EXCLUSIVE_LOCK_HELD"
	FindingInvalidIndexLeft = "INVALID_INDEX_LEFT"
	FindingSeedFailed       = "SEED_FAILED"
	FindingStatementFailed  = "STATEMENT_FAILED"
	// FindingClassUnderstated means the observed class was worse than predicted:
	// a statement static analysis called cheap turned out not to be, which is
	// the direction that reaches production.
	FindingClassUnderstated = "CLASS_UNDERSTATED"
	// FindingClassOverstated means the observed class was cheaper than predicted:
	// noise in the prediction rather than a risk in the migration.
	FindingClassOverstated = "CLASS_OVERSTATED"
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

// StatementMeasurement is one executed statement of the analyzed migration.
type StatementMeasurement struct {
	StatementIndex int               `json:"statementIndex"`
	SQL            string            `json:"sql"`
	DurationMs     int64             `json:"durationMs"`
	StrongestLock  string            `json:"strongestLock"`
	Locks          []LockObservation `json:"locks"`
	Verdict        RuntimeVerdict    `json:"verdict"`
	Error          string            `json:"error,omitempty"`
	// ObservedClass is the performance class the database produced, derived from
	// counts rather than durations. Empty when the observation is not usable: the
	// statement never completed, the run could not be wrapped in a transaction, or
	// the tables were not seeded.
	ObservedClass PerformanceClass `json:"observedClass,omitempty"`
	// PredictedClass is what static analysis said before the run, empty for a
	// statement it does not model.
	PredictedClass PerformanceClass `json:"predictedClass,omitempty"`
	// RewrittenRelations are the relations whose relfilenode changed, which is
	// true if and only if they were rewritten.
	RewrittenRelations []string `json:"rewrittenRelations"`
	// TuplesRead is how many rows the statement read from the tables it touched.
	TuplesRead int64 `json:"tuplesRead"`
}

// RuntimeFinding carries the same fields as a static analysis warning, so the
// two analysis groups stay renderable through one path.
type RuntimeFinding struct {
	Type        string      `json:"type"`
	OperationID OperationID `json:"operationId"`
	TableName   string      `json:"tableName"`
	Message     string      `json:"message"`
}

// RuntimeAnalysisResult is what runtime analysis returns: one migration executed
// against a seeded container. Its static counterpart is StaticAnalysisResult.
type RuntimeAnalysisResult struct {
	MigrationID string                 `json:"migrationId"`
	Version     string                 `json:"version"`
	DeadlineMs  int64                  `json:"deadlineMs"`
	Seeded      []SeededTable          `json:"seeded"`
	Statements  []StatementMeasurement `json:"statements"`
	Verdict     RuntimeVerdict         `json:"verdict"`
	Retry       RetryVerdict           `json:"retry"`
	Findings    []RuntimeFinding       `json:"findings"`
	Message     string                 `json:"message"`
}
