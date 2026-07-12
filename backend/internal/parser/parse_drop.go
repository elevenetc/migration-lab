package parser

import (
	pgquery "github.com/pganalyze/pg_query_go/v6"
	"migration-timeline/backend/internal/models"
)

func parseDropStmt(dropStmt *pgquery.DropStmt) []models.Operation {
	var ops []models.Operation

	if dropStmt.RemoveType != pgquery.ObjectType_OBJECT_TABLE {
		return nil
	}

	for _, obj := range dropStmt.Objects {
		list := obj.GetList()
		if list == nil {
			continue
		}

		for _, item := range list.Items {
			if s := item.GetString_(); s != nil {
				ops = append(ops, models.DropTable{
					TableName: s.Sval,
				})
			}
		}
	}

	return ops
}
