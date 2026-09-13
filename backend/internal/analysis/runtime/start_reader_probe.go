package runtime

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"migration-timeline/backend/internal/models"
)

// probeInterval leaves the migration a window to take its own locks between reads.
const probeInterval = 5 * time.Millisecond

// readerProbe is a running reader session: which backend it is, so the lock
// sampler's blocked-backend observations can be attributed to it, and how to stop
// it and collect what it saw.
type readerProbe struct {
	PID  uint32
	Stop func() models.ProbeResult
}

// startReaderProbe opens a session that reads one row of the table in a loop
// until its stop function is called. The reads take ACCESS SHARE, so a migration
// holding ACCESS EXCLUSIVE stalls them exactly as it stalls production readers.
//
// How long a read took is recorded as raw data; whether it was *blocked* is not
// inferred from that latency but read from PostgreSQL by the lock sampler, so no
// threshold has to stand in for it.
func startReaderProbe(ctx context.Context, connString, table string) (readerProbe, error) {
	conn, err := pgx.Connect(ctx, connString)
	if err != nil {
		return readerProbe{}, err
	}

	query := "SELECT 1 FROM " + quoteIdentifier(table) + " LIMIT 1"
	stopped := make(chan struct{})
	done := make(chan models.ProbeResult, 1)

	go func() {
		result := models.ProbeResult{Table: table}
		for {
			select {
			case <-stopped:
				conn.Close(context.WithoutCancel(ctx))
				done <- result
				return
			default:
			}

			start := time.Now()
			_, err := conn.Exec(ctx, query)
			latency := time.Since(start)

			result.Samples++
			if err != nil {
				result.Errors++
			}
			if latency.Milliseconds() > result.MaxLatencyMs {
				result.MaxLatencyMs = latency.Milliseconds()
			}

			time.Sleep(probeInterval)
		}
	}()

	return readerProbe{
		PID: conn.PgConn().PID(),
		Stop: func() models.ProbeResult {
			close(stopped)
			return <-done
		},
	}, nil
}
