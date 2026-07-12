package parser

import (
	pgquery "github.com/pganalyze/pg_query_go/v6"
	"migration-timeline/backend/internal/models"
)

func parseRenameStmt(renameStmt *pgquery.RenameStmt) models.Operation {
	switch renameStmt.RenameType {
	case pgquery.ObjectType_OBJECT_TABLE:
		return models.RenameTable{
			TableName:    renameStmt.Relation.Relname,
			NewTableName: renameStmt.Newname,
		}
	case pgquery.ObjectType_OBJECT_COLUMN:
		return models.RenameColumn{
			TableName:     renameStmt.Relation.Relname,
			ColumnName:    renameStmt.Subname,
			NewColumnName: renameStmt.Newname,
		}
	}
	return nil
}
