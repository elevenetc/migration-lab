package models

import "slices"

// ClassifyPerformance predicts the performance class of an operation from the
// parsed AST alone, assuming PostgreSQL 11 or newer. It is total: an operation
// this switch does not know about is reported as MetadataOnly.
//
// It is a prediction, not a verdict: runtime analysis observes the class the
// database actually produced and scores this one against it. Where the AST is
// not enough to decide, the worse class is predicted; see
// docs/supported-performance-classes.md for what stays conservative.
//
// It belongs with static analysis by subject but lives here by necessity: the
// class is injected while marshalling an operation (see MarshalOperation), so
// models cannot import the analysis packages. Injecting it rather than storing
// it also means no hand-built Migration can carry a stale or empty class.
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
		return classifyAlterColumnType(o)
	default:
		return MetadataOnly
	}
}

// A column added without a default, or with a default PostgreSQL can evaluate
// once and store in the catalog, costs nothing. A volatile default has to be
// evaluated per row, which rewrites the table. Volatility comes from the parsed
// expression tree, so a call hidden behind a cast still counts.
func classifyAddColumn(column Column) PerformanceClass {
	if !hasDefault(column) {
		return MetadataOnly
	}
	if column.DefaultVolatile {
		return TableRewrite
	}
	return MetadataOnly
}

func hasDefault(column Column) bool {
	return slices.Contains(column.Constraints, "DEFAULT")
}
