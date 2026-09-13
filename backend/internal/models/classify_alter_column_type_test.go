package models

import "testing"

// Whether a type change rewrites the table depends on both types, and the pairs
// are not symmetric: dropping a constraint is free, adding one is not.
func TestClassifyAlterColumnType(t *testing.T) {
	tests := []struct {
		name     string
		old      SQLType
		new      SQLType
		expected PerformanceClass
	}{
		{"unknown old type stays conservative", SQLType{}, NewSQLType("text"), TableRewrite},
		{"same type", NewSQLType("text"), NewSQLType("text"), MetadataOnly},
		{"same varchar length", NewSQLType("varchar", 255), NewSQLType("varchar", 255), MetadataOnly},
		{"varchar to text drops the limit", NewSQLType("varchar", 255), NewSQLType("text"), MetadataOnly},
		{"varchar to unconstrained varchar", NewSQLType("varchar", 255), NewSQLType("varchar"), MetadataOnly},
		{"widening varchar", NewSQLType("varchar", 100), NewSQLType("varchar", 255), MetadataOnly},
		{"narrowing varchar adds a constraint", NewSQLType("varchar", 255), NewSQLType("varchar", 100), TableRewrite},
		{"text to varchar adds a constraint", NewSQLType("text"), NewSQLType("varchar", 100), TableRewrite},
		{"char is blank-padded, not coercible", NewSQLType("bpchar", 10), NewSQLType("text"), TableRewrite},
		{"widening numeric precision", NewSQLType("numeric", 10, 2), NewSQLType("numeric", 12, 2), MetadataOnly},
		{"numeric to unconstrained numeric", NewSQLType("numeric", 10, 2), NewSQLType("numeric"), MetadataOnly},
		{"changing numeric scale rounds", NewSQLType("numeric", 10, 2), NewSQLType("numeric", 12, 4), TableRewrite},
		{"narrowing numeric precision", NewSQLType("numeric", 12, 2), NewSQLType("numeric", 10, 2), TableRewrite},
		{"widening integer", NewSQLType("int4"), NewSQLType("int8"), TableRewrite},
		{"unrelated types", NewSQLType("int4"), NewSQLType("text"), TableRewrite},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			op := AlterColumnType{TableName: "users", ColumnName: "email", PreviousType: test.old, NewType: test.new}
			if got := ClassifyPerformance(op); got != test.expected {
				t.Errorf("%s -> %s: expected %s, got %s", test.old, test.new, test.expected, got)
			}
		})
	}
}
