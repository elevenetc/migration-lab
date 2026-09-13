package runtime

import (
	"context"

	"migration-lab/backend/internal/models"

	"github.com/jackc/pgx/v5"
)

const invalidIndexesQuery = `SELECT count(*) FROM pg_index WHERE NOT indisvalid`

// retryVerdict classifies what a run left behind for the next attempt of the same
// migration — the question that decides whether a killed pod ever converges.
func retryVerdict(ctx context.Context, conn *pgx.Conn, verdict models.RuntimeVerdict) models.RetryVerdict {
	if verdict == models.RuntimeCompleted {
		return models.RetryNotApplicable
	}

	var invalid int
	if err := conn.QueryRow(ctx, invalidIndexesQuery).Scan(&invalid); err == nil && invalid > 0 {
		return models.RetryManualCleanup
	}

	// Nothing was left behind, so the next attempt starts from the same state,
	// does the same work, and is cancelled at the same point.
	return models.RetryFailureLoop
}
