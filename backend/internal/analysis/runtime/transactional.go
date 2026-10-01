package runtime

import (
	"strings"
)

// Statements PostgreSQL refuses to run inside a transaction block; a migration
// containing one cannot be measured and rolled back.
var nonTransactional = []string{
	"CONCURRENTLY", "VACUUM", "REINDEX", "ALTER SYSTEM",
	"CREATE DATABASE", "DROP DATABASE", "CREATE TABLESPACE", "DROP TABLESPACE",
}

// transactional reports whether the whole migration can run inside one
// transaction, which lets the measurement roll back and leave the seeded
// database as it found it.
//
// For example:
//
//	[]string{
//		"CREATE TABLE users (id INT, email TEXT)",
//		"ALTER TABLE users ADD COLUMN verified BOOLEAN DEFAULT false",
//	}
//
// is true, while
//
//	[]string{
//		"CREATE TABLE users (id INT, email TEXT)",
//		"CREATE INDEX CONCURRENTLY idx_users_email ON users (email)",
//	}
//
// is false, because one CONCURRENTLY statement keeps the whole migration out of
// a transaction.
func transactional(migrationStatements []string) bool {
	for _, statement := range migrationStatements {
		upper := strings.ToUpper(statement)
		for _, keyword := range nonTransactional {
			if strings.Contains(upper, keyword) {
				return false
			}
		}
	}
	return true
}
