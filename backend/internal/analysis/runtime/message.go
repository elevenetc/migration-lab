package runtime

import (
	"fmt"

	"migration-timeline/backend/internal/models"
)

// message is the one-line summary a report leads with: what ended the run.
func message(result models.RuntimeAnalysisResult) string {
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
