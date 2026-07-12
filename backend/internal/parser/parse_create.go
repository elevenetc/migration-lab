package parser

import (
	"fmt"
	"strings"

	pgquery "github.com/pganalyze/pg_query_go/v6"
	"migration-timeline/backend/internal/models"
)

func parseCreateStmt(createStmt *pgquery.CreateStmt) models.Operation {
	tableName := createStmt.Relation.Relname

	var columns []models.Column
	for _, elt := range createStmt.TableElts {
		colDef := elt.GetColumnDef()
		if colDef == nil {
			continue
		}

		colName := colDef.Colname
		colType := extractTypeName(colDef.TypeName)
		constraints := extractConstraints(colDef.Constraints)

		columns = append(columns, models.Column{
			Name:        colName,
			Type:        colType,
			Constraints: constraints,
		})
	}

	return models.CreateTable{
		TableName:     tableName,
		Columns:       columns,
		IsPartitioned: createStmt.Partspec != nil,
		PartitionOf:   extractPartitionOf(createStmt),
	}
}

func ParseCreateTable(sql string) (*models.CreateTable, error) {
	result, err := pgquery.Parse(sql)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SQL: %w", err)
	}

	if len(result.Stmts) == 0 {
		return nil, fmt.Errorf("no statements found")
	}

	stmt := result.Stmts[0].Stmt
	createStmt := stmt.GetCreateStmt()
	if createStmt == nil {
		return nil, fmt.Errorf("not a CREATE TABLE statement")
	}

	op := parseCreateStmt(createStmt)
	ct := op.(models.CreateTable)
	return &ct, nil
}

func extractTypeName(typeName *pgquery.TypeName) string {
	if typeName == nil {
		return ""
	}

	var parts []string
	for _, name := range typeName.Names {
		str := name.GetString_()
		if str != nil && str.Sval != "pg_catalog" {
			parts = append(parts, str.Sval)
		}
	}

	result := strings.Join(parts, ".")

	if len(typeName.Typmods) > 0 {
		var mods []string
		for _, mod := range typeName.Typmods {
			if constVal := mod.GetAConst(); constVal != nil {
				if ival := constVal.GetIval(); ival != nil {
					mods = append(mods, fmt.Sprintf("%d", ival.Ival))
				}
			}
		}
		if len(mods) > 0 {
			result = fmt.Sprintf("%s(%s)", result, strings.Join(mods, ","))
		}
	}

	return result
}

func extractConstraints(constraints []*pgquery.Node) []string {
	var result []string
	for _, c := range constraints {
		constraint := c.GetConstraint()
		if constraint == nil {
			continue
		}

		switch constraint.Contype {
		case pgquery.ConstrType_CONSTR_PRIMARY:
			result = append(result, "PRIMARY KEY")
		case pgquery.ConstrType_CONSTR_NOTNULL:
			result = append(result, "NOT NULL")
		case pgquery.ConstrType_CONSTR_UNIQUE:
			result = append(result, "UNIQUE")
		case pgquery.ConstrType_CONSTR_DEFAULT:
			result = append(result, "DEFAULT")
		case pgquery.ConstrType_CONSTR_NULL:
			result = append(result, "NULL")
		}
	}
	return result
}

func extractPartitionOf(createStmt *pgquery.CreateStmt) *string {
	if createStmt.InhRelations == nil || len(createStmt.InhRelations) == 0 {
		return nil
	}
	if createStmt.Partbound != nil {
		rangeVar := createStmt.InhRelations[0].GetRangeVar()
		if rangeVar != nil {
			return &rangeVar.Relname
		}
	}
	return nil
}
