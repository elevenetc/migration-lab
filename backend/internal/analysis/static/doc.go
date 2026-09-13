// Package static raises warnings that can be read off the parsed AST, without a
// database: anything derivable from the migration text belongs here.
//
// It is one of the two analysis groups under internal/analysis; the other,
// runtime, executes a migration and reports what it measured. Static analysis
// also predicts the performance class of an operation, but that prediction lives
// in models, next to the operations, because it is injected at marshal time —
// see models/classify_performance.go.
package static
