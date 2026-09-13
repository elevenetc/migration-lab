package datasets

import "migration-lab/backend/internal/models"

type migration struct {
	version string
	sql     string
}

// Datasets returns the built-in migration datasets keyed by id, as raw sources
// (ID/SQL/Timestamp) for the temporary Database to parse and run.
func Datasets() map[string][]models.MigrationInfo {
	return map[string][]models.MigrationInfo{
		"ecommerce":                        toInfos(complexEcommerceMigrations),
		"simple-partition":                 toInfos(simplePartitionMigrations),
		"simple-rename":                    toInfos(simpleRenameMigrations),
		"rename-with-new-table-in-between": toInfos(renameWithNewTableInBetweenMigrations),
		"simple-table":                     toInfos(simpleTableMigrations),
		"create-add-column-rename":         toInfos(createAddColumnRenameMigrations),
		"multiple-operation-per-migration": toInfos(multipleOperationPerMigrationMigrations),
		"warning-partition-parent-alter":   toInfos(warningPartitionParentAlterMigrations),
		"performance-class":                toInfos(performanceClassMigrations),
		"tall":                             toInfos(tallMigrations),
		"wide":                             toInfos(wideMigrations),
	}
}

func toInfos(migrations []migration) []models.MigrationInfo {
	infos := make([]models.MigrationInfo, len(migrations))
	for i, m := range migrations {
		infos[i] = models.MigrationInfo{ID: m.version, SQL: m.sql, Timestamp: int64(i + 1)}
	}
	return infos
}
