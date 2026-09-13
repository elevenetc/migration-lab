package runtime

import (
	"context"
	"fmt"
	"log"

	"migration-lab/backend/internal/models"

	"github.com/jackc/pgx/v5"
)

// Seed carries out every plan to the given row count, then analyzes the table so
// the planner costs the migration against real statistics. A table that cannot be
// filled is reported with its error instead of failing the run.
func Seed(ctx context.Context, conn *pgx.Conn, plans []SeedPlan, rows int64) []models.SeededTable {
	seeded := make([]models.SeededTable, 0, len(plans))

	for _, plan := range plans {
		statement := SeedStatement(plan, rows)
		if statement == "" {
			seeded = append(seeded, models.SeededTable{
				Table: plan.Table,
				Error: "no column of the table can be filled generically",
			})
			continue
		}

		log.Printf("Seeding %s with %d rows", plan.Table, rows)
		if _, err := conn.Exec(ctx, statement); err != nil {
			seeded = append(seeded, models.SeededTable{Table: plan.Table, Error: err.Error()})
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
