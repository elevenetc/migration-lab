package parser

import (
	pgquery "github.com/pganalyze/pg_query_go/v6"
	"migration-timeline/backend/internal/models"
)

func parseAlterTableStmt(alterStmt *pgquery.AlterTableStmt) []models.Operation {
	var ops []models.Operation
	tableName := alterStmt.Relation.Relname

	for _, cmd := range alterStmt.Cmds {
		alterCmd := cmd.GetAlterTableCmd()
		if alterCmd == nil {
			continue
		}

		switch alterCmd.Subtype {
		case pgquery.AlterTableType_AT_AddColumn:
			if op := parseAddColumn(tableName, alterCmd); op != nil {
				ops = append(ops, op)
			}
		case pgquery.AlterTableType_AT_DropColumn:
			ops = append(ops, models.DropColumn{
				TableName:  tableName,
				ColumnName: alterCmd.Name,
			})
		case pgquery.AlterTableType_AT_AlterColumnType:
			if op := parseAlterColumnType(tableName, alterCmd); op != nil {
				ops = append(ops, op)
			}
		case pgquery.AlterTableType_AT_SetNotNull:
			ops = append(ops, models.SetNotNull{
				TableName:  tableName,
				ColumnName: alterCmd.Name,
			})
		case pgquery.AlterTableType_AT_DropNotNull:
			ops = append(ops, models.DropNotNull{
				TableName:  tableName,
				ColumnName: alterCmd.Name,
			})
		case pgquery.AlterTableType_AT_ColumnDefault:
			if alterCmd.Def != nil {
				ops = append(ops, parseSetDefault(tableName, alterCmd))
			} else {
				ops = append(ops, models.DropDefault{
					TableName:  tableName,
					ColumnName: alterCmd.Name,
				})
			}
		case pgquery.AlterTableType_AT_AddConstraint:
			if op := parseAddConstraint(tableName, alterCmd); op != nil {
				ops = append(ops, op)
			}
		case pgquery.AlterTableType_AT_DropConstraint:
			ops = append(ops, models.DropConstraint{
				TableName:      tableName,
				ConstraintName: alterCmd.Name,
			})
		}
	}

	return ops
}

func parseAddColumn(tableName string, cmd *pgquery.AlterTableCmd) models.Operation {
	def := cmd.Def
	if def == nil {
		return nil
	}

	colDef := def.GetColumnDef()
	if colDef == nil {
		return nil
	}

	return models.AddColumn{
		TableName: tableName,
		Column: models.Column{
			Name:        colDef.Colname,
			Type:        extractTypeName(colDef.TypeName),
			Constraints: extractConstraints(colDef.Constraints),
		},
	}
}

func parseAlterColumnType(tableName string, cmd *pgquery.AlterTableCmd) models.Operation {
	def := cmd.Def
	if def == nil {
		return nil
	}

	colDef := def.GetColumnDef()
	if colDef == nil {
		return nil
	}

	return models.AlterColumnType{
		TableName:  tableName,
		ColumnName: cmd.Name,
		NewType:    extractTypeName(colDef.TypeName),
	}
}

func parseSetDefault(tableName string, cmd *pgquery.AlterTableCmd) models.Operation {
	defaultValue := deparseDef(cmd.Def)
	return models.SetDefault{
		TableName:    tableName,
		ColumnName:   cmd.Name,
		DefaultValue: defaultValue,
	}
}

func parseAddConstraint(tableName string, cmd *pgquery.AlterTableCmd) models.Operation {
	def := cmd.Def
	if def == nil {
		return nil
	}

	constraint := def.GetConstraint()
	if constraint == nil {
		return nil
	}

	constraintType := mapConstraintType(constraint.Contype)
	return models.AddConstraint{
		TableName:      tableName,
		ConstraintName: constraint.Conname,
		ConstraintType: constraintType,
	}
}

func mapConstraintType(contype pgquery.ConstrType) string {
	switch contype {
	case pgquery.ConstrType_CONSTR_PRIMARY:
		return "PRIMARY_KEY"
	case pgquery.ConstrType_CONSTR_UNIQUE:
		return "UNIQUE"
	case pgquery.ConstrType_CONSTR_FOREIGN:
		return "FOREIGN_KEY"
	case pgquery.ConstrType_CONSTR_CHECK:
		return "CHECK"
	case pgquery.ConstrType_CONSTR_EXCLUSION:
		return "EXCLUSION"
	default:
		return "UNKNOWN"
	}
}

func deparseDef(node *pgquery.Node) string {
	if node == nil {
		return ""
	}

	if aConst := node.GetAConst(); aConst != nil {
		if sval := aConst.GetSval(); sval != nil {
			return "'" + sval.Sval + "'"
		}
		if ival := aConst.GetIval(); ival != nil {
			return string(rune(ival.Ival))
		}
		if boolval := aConst.GetBoolval(); boolval != nil {
			if boolval.Boolval {
				return "true"
			}
			return "false"
		}
	}

	if funcCall := node.GetFuncCall(); funcCall != nil {
		return deparseFuncCall(funcCall)
	}

	if typeCast := node.GetTypeCast(); typeCast != nil {
		return deparseTypeCast(typeCast)
	}

	return ""
}

func deparseFuncCall(fc *pgquery.FuncCall) string {
	if fc == nil || len(fc.Funcname) == 0 {
		return ""
	}

	var funcName string
	for _, n := range fc.Funcname {
		if s := n.GetString_(); s != nil {
			if funcName != "" {
				funcName += "."
			}
			funcName += s.Sval
		}
	}

	return funcName + "()"
}

func deparseTypeCast(tc *pgquery.TypeCast) string {
	if tc == nil {
		return ""
	}

	arg := deparseDef(tc.Arg)
	typeName := extractTypeName(tc.TypeName)

	if arg != "" && typeName != "" {
		return arg + "::" + typeName
	}
	if arg != "" {
		return arg
	}
	return ""
}
