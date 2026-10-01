package runtime

import (
	"context"
	"fmt"

	"migration-lab/backend/internal/monitoring"
	"migration-lab/backend/internal/pg"

	"github.com/jackc/pgx/v5"
)

// startSession starts a PostgreSQL container and opens the two connections a
// measurement needs: one running the migrations, and one sampling the locks it
// holds. stop closes both and terminates the container; it is a no-op when
// startSession fails.
func startSession(ctx context.Context) (conn, sampler *pgx.Conn, stop func(), err error) {
	connString, terminate, err := pg.Start(ctx)
	if err != nil {
		return nil, nil, func() {}, err
	}

	conn, err = pgx.Connect(ctx, connString)
	if err != nil {
		monitoring.CaptureError(ctx, err)
		terminate()
		return nil, nil, func() {}, fmt.Errorf("failed to connect to the database: %w", err)
	}

	sampler, err = pgx.Connect(ctx, connString)
	if err != nil {
		monitoring.CaptureError(ctx, err)
		closeConn(ctx, conn, "runtime")
		terminate()
		return nil, nil, func() {}, fmt.Errorf("failed to open a sampling connection: %w", err)
	}

	return conn, sampler, func() {
		closeConn(ctx, sampler, "lock sampler")
		closeConn(ctx, conn, "runtime")
		terminate()
	}, nil
}

func closeConn(ctx context.Context, conn *pgx.Conn, name string) {
	if err := conn.Close(context.WithoutCancel(ctx)); err != nil {
		monitoring.Errorf(ctx, "Failed to close %s connection: %v", name, err)
	}
}
