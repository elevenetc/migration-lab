package models

// PerformanceClass is how expensive an operation is relative to table size.
// The classes are ordinal: MetadataOnly < DataScanning < TableRewrite.
type PerformanceClass string

const (
	// MetadataOnly changes the catalog only; cost is constant in table size.
	MetadataOnly PerformanceClass = "METADATA_ONLY"
	// DataScanning reads every existing row to validate it or to build an index.
	DataScanning PerformanceClass = "DATA_SCANNING"
	// TableRewrite rewrites the whole table on disk.
	TableRewrite PerformanceClass = "TABLE_REWRITE"
)

func performanceRank(class PerformanceClass) int {
	switch class {
	case TableRewrite:
		return 2
	case DataScanning:
		return 1
	default:
		return 0
	}
}
