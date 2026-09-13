package runtime

import (
	"context"
	"log"

	"migration-timeline/backend/internal/models"
)

// maxProbes bounds the concurrent reader sessions opened alongside the migration.
const maxProbes = 4

// startProbes opens a reader session per seeded table and returns the function
// that stops them all. Stopping takes the sampler's blocked-backend tick counts,
// which is where each probe's waiting time comes from.
func startProbes(ctx context.Context, connString string, tables []string) func(map[uint32]int) []models.ProbeResult {
	var probes []readerProbe

	for _, table := range tables {
		probe, err := startReaderProbe(ctx, connString, table)
		if err != nil {
			log.Printf("Failed to start a reader probe on %s: %v", table, err)
			continue
		}
		probes = append(probes, probe)
	}

	return func(blockedTicks map[uint32]int) []models.ProbeResult {
		results := make([]models.ProbeResult, 0, len(probes))
		for _, probe := range probes {
			result := probe.Stop()
			result.BlockedMs = int64(blockedTicks[probe.PID]) * lockSampleInterval.Milliseconds()
			results = append(results, result)
		}
		return results
	}
}

// probedTables are the tables worth reading concurrently: the ones that actually
// hold rows, capped so a wide migration does not open a session per table.
func probedTables(seeded []models.SeededTable) []string {
	var tables []string
	for _, table := range seeded {
		if table.Error != "" || table.Rows == 0 {
			continue
		}
		if len(tables) == maxProbes {
			break
		}
		tables = append(tables, table.Table)
	}
	return tables
}
