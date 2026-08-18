package models

import "slices"

import "strings"

// ClassifyPerformance derives the performance class of an operation from the
// parsed AST alone, assuming PostgreSQL 11 or newer. It is total: an operation
// this switch does not know about is reported as MetadataOnly.
//
// Context-dependent cases are classified to the worse class rather than parsed
// in full; see docs/supported-performance-classes.md for the assumptions.
func ClassifyPerformance(op Operation) PerformanceClass {
	switch o := op.(type) {
	case AddColumn:
		return classifyAddColumn(o.Column)
	case AddConstraint:
		if o.NotValid {
			return MetadataOnly
		}
		return DataScanning
	case SetNotNull:
		return DataScanning
	case AlterColumnType:
		return TableRewrite
	default:
		return MetadataOnly
	}
}

// A column added without a default, or with a default PostgreSQL can evaluate
// once and store in the catalog, costs nothing. A function-call default is
// treated as volatile - and so as a rewrite - even though PostgreSQL stores
// non-volatile ones such as now() without touching existing rows. A default we
// failed to deparse takes the same conservative path.
func classifyAddColumn(column Column) PerformanceClass {
	if !hasDefault(column) {
		return MetadataOnly
	}
	if column.DefaultExpr == "" || strings.HasSuffix(column.DefaultExpr, "()") {
		return TableRewrite
	}
	return MetadataOnly
}

func hasDefault(column Column) bool {
	return slices.Contains(column.Constraints, "DEFAULT")
}
