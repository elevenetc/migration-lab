package models

// WorstPerformanceClass is the class of a whole statement: the most expensive
// class among its operations, MetadataOnly when it has none.
func WorstPerformanceClass(ops []Operation) PerformanceClass {
	worst := MetadataOnly
	for _, op := range ops {
		if class := ClassifyPerformance(op); performanceRank(class) > performanceRank(worst) {
			worst = class
		}
	}
	return worst
}
