package parser

import (
	"testing"

	"migration-lab/backend/internal/models"
)

// pg_query canonicalises the spelling of every built-in type, so the classifier
// downstream never has to know the aliases a migration may have written.
func TestExtractColumnTypeCanonicalisesDeclaredTypes(t *testing.T) {
	tests := []struct {
		declared string
		expected models.SQLType
	}{
		{"CHARACTER VARYING(255)", models.NewSQLType("varchar", 255)},
		{"VARCHAR", models.NewSQLType("varchar")},
		{"CHAR(3)", models.NewSQLType("bpchar", 3)},
		{"DECIMAL(10,2)", models.NewSQLType("numeric", 10, 2)},
		{"INT", models.NewSQLType("int4")},
		{"BIGINT", models.NewSQLType("int8")},
		{"TEXT", models.NewSQLType("text")},
		{"citext", models.NewSQLType("citext")},
	}

	for _, test := range tests {
		t.Run(test.declared, func(t *testing.T) {
			table, err := ParseCreateTable("CREATE TABLE t (c " + test.declared + ");")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			got := table.Columns[0].Type
			if got.String() != test.expected.String() {
				t.Errorf("expected %s, got %s", test.expected, got)
			}
		})
	}
}
