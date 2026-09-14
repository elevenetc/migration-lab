// Package runtime measures one migration against a seeded PostgreSQL container
// and raises findings from what it observed: durations, locks,
// leftovers and the performance class the database actually produced.
//
// It is one of the two analysis groups under internal/analysis; the other,
// static, needs no database. Anything derivable from the AST belongs there
// instead. Analyse is the only entry point, and everything that talks to the
// database is confined to its own file.
package runtime
