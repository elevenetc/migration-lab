package runtime

import (
	"testing"

	"migration-lab/backend/internal/models"
)

func addColumnMigration() *models.Migration {
	return &models.Migration{
		ID: "V7__add_created_at",
		Statements: []models.Statement{{
			Index: 0,
			SQL:   "ALTER TABLE accounts ADD COLUMN created_at TIMESTAMP DEFAULT now()",
			Operations: []models.Operation{models.AddColumn{
				TableName: "accounts",
				Column:    models.Column{Name: "created_at", Type: models.NewSQLType("timestamp"), Constraints: []string{"DEFAULT"}},
			}},
		}},
	}
}

func TestWithClassesPairsTheMeasurementWithThePrediction(t *testing.T) {
	seeded := []models.SeededTable{{Table: "accounts", Rows: 1_000_000}}
	measurements := []models.StatementMeasurement{{
		SQL:           "ALTER TABLE accounts ADD COLUMN created_at TIMESTAMP DEFAULT now()",
		ObservedClass: models.MetadataOnly,
		Verdict:       models.RuntimeCompleted,
	}}

	scored := withClasses(addColumnMigration(), seeded, measurements)

	if scored[0].PredictedClass != models.MetadataOnly {
		t.Errorf("predicted = %q, want %q", scored[0].PredictedClass, models.MetadataOnly)
	}
	if scored[0].ObservedClass != models.MetadataOnly {
		t.Errorf("observed = %q, want it kept", scored[0].ObservedClass)
	}
	if measurements[0].PredictedClass != "" {
		t.Error("expected the given measurements to be left alone")
	}
}

func TestWithClassesLeavesAnUnmodelledStatementUnpredicted(t *testing.T) {
	measurements := []models.StatementMeasurement{{SQL: "VACUUM FULL accounts", ObservedClass: models.TableRewrite}}

	scored := withClasses(addColumnMigration(), nil, measurements)

	if scored[0].PredictedClass != "" {
		t.Errorf("predicted = %q, want nothing for a statement static analysis does not model", scored[0].PredictedClass)
	}
}

// A table seeded to nothing reads no rows and rewrites nothing measurable, which
// is the same shape a genuinely cheap statement produces. Blind must not read as
// clean.
func TestWithClassesDiscardsACheapObservationOfAnEmptyTable(t *testing.T) {
	seeded := []models.SeededTable{{Table: "accounts", Error: "no column of the table can be filled generically"}}
	measurements := []models.StatementMeasurement{{
		SQL:           "ALTER TABLE accounts ADD COLUMN created_at TIMESTAMP DEFAULT now()",
		ObservedClass: models.MetadataOnly,
	}}

	scored := withClasses(addColumnMigration(), seeded, measurements)

	if scored[0].ObservedClass != "" {
		t.Errorf("observed = %q, want it discarded so the prediction stands in", scored[0].ObservedClass)
	}
	if scored[0].PredictedClass != models.MetadataOnly {
		t.Errorf("predicted = %q, want the prediction kept", scored[0].PredictedClass)
	}
}

// A changed relfilenode proves a rewrite at any row count, so that observation
// survives a failed seed.
func TestWithClassesKeepsAnObservedRewriteOfAnEmptyTable(t *testing.T) {
	seeded := []models.SeededTable{{Table: "accounts", Rows: 0}}
	measurements := []models.StatementMeasurement{{
		SQL:                "ALTER TABLE accounts ADD COLUMN created_at TIMESTAMP DEFAULT now()",
		ObservedClass:      models.TableRewrite,
		RewrittenRelations: []string{"accounts"},
	}}

	scored := withClasses(addColumnMigration(), seeded, measurements)

	if scored[0].ObservedClass != models.TableRewrite {
		t.Errorf("observed = %q, want %q", scored[0].ObservedClass, models.TableRewrite)
	}
}
