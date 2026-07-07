package analysis

import (
	"fmt"

	"migration-timeline/backend/internal/models"
)

func detectAccessExclusiveLock(op models.Operation, ctx *analysisContext) models.Warning {
	var tableName, migrationID string

	switch o := op.(type) {
	case models.AddColumn:
		tableName, migrationID = o.TableName, o.MigrationID
	case models.DropColumn:
		tableName, migrationID = o.TableName, o.MigrationID
	case models.AlterColumnType:
		tableName, migrationID = o.TableName, o.MigrationID
	case models.SetNotNull:
		tableName, migrationID = o.TableName, o.MigrationID
	case models.DropNotNull:
		tableName, migrationID = o.TableName, o.MigrationID
	case models.SetDefault:
		tableName, migrationID = o.TableName, o.MigrationID
	case models.DropDefault:
		tableName, migrationID = o.TableName, o.MigrationID
	case models.AddConstraint:
		tableName, migrationID = o.TableName, o.MigrationID
	case models.DropConstraint:
		tableName, migrationID = o.TableName, o.MigrationID
	case models.RenameColumn:
		tableName, migrationID = o.TableName, o.MigrationID
	default:
		return nil
	}

	if !ctx.partitionedTables[tableName] {
		return nil
	}

	return models.AccessExclusiveLock{
		OperationID: models.OperationID{
			MigrationID: migrationID,
			TableName:   tableName,
		},
		TableName: tableName,
		Message:   fmt.Sprintf("ALTER on partitioned table '%s' acquires ACCESS EXCLUSIVE lock on parent and all partitions", tableName),
	}
}
