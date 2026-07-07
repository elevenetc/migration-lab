package models

type MigrationInfo struct {
	ID        string
	SQL       string
	Timestamp int64
}

type RunMigrationsResult struct {
	Success           bool   `json:"success"`
	Message           string `json:"message"`
	MigrationsApplied int    `json:"migrationsApplied"`
}
