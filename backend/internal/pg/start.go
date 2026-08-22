package pg

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Image is pinned so version-gated behaviour (the SET NOT NULL scan skip, the
// metadata-only ADD COLUMN ... DEFAULT fast path) matches what is measured.
const Image = "postgres:16-alpine"

// Start boots a throwaway PostgreSQL container and returns its connection
// string together with the function that tears it down.
func Start(ctx context.Context) (string, func(), error) {
	container, err := postgres.Run(ctx,
		Image,
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		return "", func() {}, fmt.Errorf("failed to start PostgreSQL container: %w", err)
	}

	terminate := func() {
		if err := container.Terminate(context.WithoutCancel(ctx)); err != nil {
			log.Printf("Failed to terminate container: %v", err)
		}
	}

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		terminate()
		return "", func() {}, fmt.Errorf("failed to get connection string: %w", err)
	}

	return connStr, terminate, nil
}
