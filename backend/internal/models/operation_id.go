package models

// StatementScoped marks a warning that applies to a whole statement rather
// than a single operation within it.
const StatementScoped = -1

// OperationID addresses one operation in the timeline. Both analysis groups use
// it — a static Warning and a RuntimeFinding are anchored the same way — so it
// belongs to neither of them.
type OperationID struct {
	MigrationID    string `json:"migrationId"`
	StatementIndex int    `json:"statementIndex"`
	OpIndex        int    `json:"opIndex"` // StatementScoped (-1) for statement-level warnings
}
