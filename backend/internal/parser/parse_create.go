package parser

import (
	"fmt"

	pgquery "github.com/pganalyze/pg_query_go/v6"
	"migration-lab/backend/internal/models"
)

func parseCreateStmt(createStmt *pgquery.CreateStmt) models.Operation {
	tableName := createStmt.Relation.Relname

	var columns []models.Column
	for _, elt := range createStmt.TableElts {
		colDef := elt.GetColumnDef()
		if colDef == nil {
			continue
		}

		columns = append(columns, models.Column{
			Name:            colDef.Colname,
			Type:            extractColumnType(colDef.TypeName),
			Constraints:     extractConstraints(colDef.Constraints),
			DefaultExpr:     extractDefaultExpr(colDef.Constraints),
			DefaultVolatile: extractDefaultVolatility(colDef.Constraints),
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
