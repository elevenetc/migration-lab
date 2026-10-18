package runtime

import (
	"context"
	"fmt"

	"migration-lab/backend/internal/models"
	"migration-lab/backend/internal/monitoring"

	"github.com/jackc/pgx/v5"
)

// Seed tops every planned table up to the given row count, then analyzes it so
// the planner costs the migration against real statistics. Topping up rather
// than appending keeps a table seeded for an earlier migration of a timeline at
// the same size, and keeps rows a migration inserted itself. A table that cannot
// be filled is reported with its error instead of failing the run.
func Seed(ctx context.Context, conn *pgx.Conn, plans []SeedPlan, rows int64) []models.SeededTable {
	seeded := make([]models.SeededTable, 0, len(plans))

	for _, plan := range plans {
		var existing int64
		if err := conn.QueryRow(ctx, "SELECT count(*) FROM "+quoteIdentifier(plan.Table)).Scan(&existing); err != nil {
			seeded = append(seeded, models.SeededTable{Table: plan.Table, Error: err.Error()})
			continue
		}
		missing := missingRows(existing, rows)
		if missing == 0 {
			seeded = append(seeded, models.SeededTable{Table: plan.Table, Rows: existing})
			continue
		}

		monitoring.Infof(ctx, "Seeding %s with %d rows", plan.Table, missing)
		if _, err := conn.Exec(ctx, SeedStatement(plan, existing+1, existing+missing)); err != nil {
			seeded = append(seeded, models.SeededTable{Table: plan.Table, Rows: existing, Error: err.Error()})
			continue
		}
		if _, err := conn.Exec(ctx, "ANALYZE "+quoteIdentifier(plan.Table)); err != nil {
			seeded = append(seeded, models.SeededTable{
				Table: plan.Table,
				Rows:  rows,
				Error: fmt.Sprintf("seeded but ANALYZE failed: %v", err),
			})
			continue
		}
		seeded = append(seeded, models.SeededTable{Table: plan.Table, Rows: rows})
	}
	return seeded
}

// missingRows is how many rows bring a table holding existing rows up to target.
func missingRows(existing, target int64) int64 {
	return max(target-existing, 0)
}
