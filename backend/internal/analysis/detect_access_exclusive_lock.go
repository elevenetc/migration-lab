package analysis

import (
	"fmt"

	"migration-timeline/backend/internal/models"
)

// The ACCESS EXCLUSIVE lock is acquired per statement, so the warning is
// statement-scoped: one warning even if the statement holds several ALTER
// commands on the table.
func detectAccessExclusiveLock(stmt models.Statement, migrationID string, ctx *analysisContext) models.Warning {
	for _, op := range stmt.Operations {
		tableName := alteredTableName(op)
		if tableName == "" || !ctx.partitionedTables[tableName] {
			continue
		}

		return models.AccessExclusiveLock{
			OperationID: models.OperationID{
				MigrationID:    migrationID,
				StatementIndex: stmt.Index,
				OpIndex:        models.StatementScoped,
			},
			TableName: tableName,
			Message:   fmt.Sprintf("ALTER on partitioned table '%s' acquires ACCESS EXCLUSIVE lock on parent and all partitions", tableName),
		}
	}
	return nil
}

// alteredTableName returns the table an ALTER-derived operation targets,
// or "" for operations that do not take the ACCESS EXCLUSIVE lock path.
func alteredTableName(op models.Operation) string {
	switch o := op.(type) {
	case models.AddColumn:
		return o.TableName
	case models.DropColumn:
		return o.TableName
	case models.AlterColumnType:
		return o.TableName
	case models.SetNotNull:
		return o.TableName
	case models.DropNotNull:
		return o.TableName
	case models.SetDefault:
		return o.TableName
	case models.DropDefault:
		return o.TableName
	case models.AddConstraint:
		return o.TableName
	case models.DropConstraint:
		return o.TableName
	case models.RenameColumn:
		return o.TableName
	default:
		return ""
	}
}
