package runtime

import (
	"migration-lab/backend/internal/models"
)

// withClasses pairs every measurement with what static analysis predicted for it,
// and drops an observation the seed made worthless. It returns new measurements
// rather than editing the ones it was given.
func withClasses(migration *models.Migration, seeded []models.SeededTable, measurements []models.StatementMeasurement) []models.StatementMeasurement {
	trusted := seedTrusted(seeded)
	scored := make([]models.StatementMeasurement, 0, len(measurements))

	for _, measurement := range measurements {
		if class, known := classOfSQL(migration, measurement.SQL); known {
			measurement.PredictedClass = class
		}
		if !trusted {
			measurement.ObservedClass = observedWithoutRows(measurement.ObservedClass)
		}
		scored = append(scored, measurement)
	}

	return scored
}

// seedTrusted reports whether the run measured tables that actually held rows.
//
// A table seeded to nothing reads no rows and rewrites nothing measurable, so the
// observation collapses to METADATA_ONLY — the same shape a genuinely cheap
// statement produces. Trust is a property of the whole run rather than of one
// table, because a partitioned parent is seeded through leaves whose names do not
// match the statement's table; SEED_FAILED already names the table at fault.
func seedTrusted(seeded []models.SeededTable) bool {
	for _, table := range seeded {
		if table.Error != "" || table.Rows == 0 {
			return false
		}
	}
	return true
}

// observedWithoutRows is what an observation of unseeded tables still proves. A
// changed relfilenode is a rewrite whatever the row count; the two cheaper
// classes are indistinguishable from having measured nothing.
func observedWithoutRows(class models.PerformanceClass) models.PerformanceClass {
	if class == models.TableRewrite {
		return models.TableRewrite
	}
	return ""
}
