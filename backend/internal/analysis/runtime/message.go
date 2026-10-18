package runtime

import (
	"fmt"

	"migration-lab/backend/internal/models"
)

// message is the one-line summary a report leads with: what ended the run.
func message(result models.RuntimeMigrationResult) string {
	if len(result.Statements) == 0 {
		return "the migration has no statements to run"
	}

	last := result.Statements[len(result.Statements)-1]

	switch last.Verdict {
	case models.RuntimeExceedsDeadline:
		return fmt.Sprintf("statement %d was cancelled after the %d ms deadline",
			last.StatementIndex, result.DeadlineMs)
	case models.RuntimeFailed:
		return fmt.Sprintf("statement %d failed: %s", last.StatementIndex, last.Error)
	default:
		return fmt.Sprintf("all %d statements completed in %d ms",
			len(result.Statements), totalDurationMs(result.Statements))
	}
}

func totalDurationMs(measurements []models.StatementMeasurement) int64 {
	var total int64
	for _, measurement := range measurements {
		total += measurement.DurationMs
	}
	return total
}

// analysisMessage is the one-line summary of every measured migration.
func analysisMessage(migrations []models.RuntimeMigrationResult) string {
	var total int64
	incomplete := 0
	for _, migration := range migrations {
		total += totalDurationMs(migration.Statements)
		if migration.Verdict != models.RuntimeCompleted {
			incomplete++
		}
	}
	if incomplete == 0 {
		return fmt.Sprintf("all %d migrations completed in %d ms", len(migrations), total)
	}
	return fmt.Sprintf("%d of %d migrations did not complete", incomplete, len(migrations))
}
