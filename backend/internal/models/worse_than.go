package models

// WorseThan reports whether class a is more expensive than class b, so a
// prediction and an observation can be compared outside this package.
func WorseThan(a, b PerformanceClass) bool {
	return performanceRank(a) > performanceRank(b)
}
