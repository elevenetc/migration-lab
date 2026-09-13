package models

import "testing"

func TestClassifyPerformance(t *testing.T) {
	tests := []struct {
		name     string
		op       Operation
		expected PerformanceClass
	}{
		{"create table", CreateTable{TableName: "users"}, MetadataOnly},
		{"drop table", DropTable{TableName: "users"}, MetadataOnly},
		{"drop column", DropColumn{TableName: "users", ColumnName: "email"}, MetadataOnly},
		{"rename table", RenameTable{TableName: "users", NewTableName: "accounts"}, MetadataOnly},
		{"rename column", RenameColumn{TableName: "users", ColumnName: "email"}, MetadataOnly},
		{"set default", SetDefault{TableName: "users", ColumnName: "status"}, MetadataOnly},
		{"drop default", DropDefault{TableName: "users", ColumnName: "status"}, MetadataOnly},
		{"drop not null", DropNotNull{TableName: "users", ColumnName: "email"}, MetadataOnly},
		{"drop constraint", DropConstraint{TableName: "users", ConstraintName: "uk"}, MetadataOnly},

		{
			"add nullable column",
			AddColumn{TableName: "users", Column: Column{Name: "note", Type: NewSQLType("text")}},
			MetadataOnly,
		},
		{
			"add column with constant default",
			AddColumn{TableName: "users", Column: Column{
				Name: "status", Type: NewSQLType("text"), Constraints: []string{"DEFAULT"}, DefaultExpr: "'active'",
			}},
			MetadataOnly,
		},
		{
			"add column with non-volatile function default",
			AddColumn{TableName: "users", Column: Column{
				Name: "created_at", Type: NewSQLType("timestamp"), Constraints: []string{"DEFAULT"}, DefaultExpr: "now()",
			}},
			MetadataOnly,
		},
		{
			"add column with volatile default",
			AddColumn{TableName: "users", Column: Column{
				Name: "token", Type: NewSQLType("int4"), Constraints: []string{"DEFAULT"},
				DefaultExpr: "random()::int4", DefaultVolatile: true,
			}},
			TableRewrite,
		},

		{"set not null", SetNotNull{TableName: "users", ColumnName: "email"}, DataScanning},
		{
			"add validating constraint",
			AddConstraint{TableName: "users", ConstraintName: "ck", ConstraintType: "CHECK"},
			DataScanning,
		},
		{
			"add not valid constraint",
			AddConstraint{TableName: "users", ConstraintName: "ck", ConstraintType: "CHECK", NotValid: true},
			MetadataOnly,
		},

		{
			"alter column type from an unknown type",
			AlterColumnType{TableName: "users", ColumnName: "email", NewType: NewSQLType("text")},
			TableRewrite,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := ClassifyPerformance(test.op); got != test.expected {
				t.Errorf("expected %s, got %s", test.expected, got)
			}
		})
	}
}

func TestWorstPerformanceClass(t *testing.T) {
	tests := []struct {
		name     string
		ops      []Operation
		expected PerformanceClass
	}{
		{"no operations", nil, MetadataOnly},
		{
			"all metadata only",
			[]Operation{DropColumn{TableName: "users"}, DropNotNull{TableName: "users"}},
			MetadataOnly,
		},
		{
			"scanning beats metadata only",
			[]Operation{DropColumn{TableName: "users"}, SetNotNull{TableName: "users"}},
			DataScanning,
		},
		{
			"rewrite beats scanning",
			[]Operation{SetNotNull{TableName: "users"}, AlterColumnType{TableName: "users"}},
			TableRewrite,
		},
		{
			"order does not matter",
			[]Operation{AlterColumnType{TableName: "users"}, SetNotNull{TableName: "users"}},
			TableRewrite,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := WorstPerformanceClass(test.ops); got != test.expected {
				t.Errorf("expected %s, got %s", test.expected, got)
			}
		})
	}
}
