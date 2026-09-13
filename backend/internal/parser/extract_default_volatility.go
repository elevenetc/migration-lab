package parser

import (
	"slices"
	"strings"

	pgquery "github.com/pganalyze/pg_query_go/v6"
)

// Functions PostgreSQL evaluates once for the whole statement, so a DEFAULT
// built from them is stored in the catalog instead of written to every row.
// Anything absent from this set counts as volatile, which keeps an unknown
// function on the conservative side.
var nonVolatileFunctions = map[string]bool{
	"now":                   true,
	"current_timestamp":     true,
	"current_date":          true,
	"current_time":          true,
	"localtime":             true,
	"localtimestamp":        true,
	"statement_timestamp":   true,
	"transaction_timestamp": true,
	"current_user":          true,
	"session_user":          true,
	"current_schema":        true,
	"upper":                 true,
	"lower":                 true,
	"concat":                true,
	"md5":                   true,
	"abs":                   true,
	"length":                true,
	"coalesce":              true,
}

// extractDefaultVolatility reports whether a column's DEFAULT may be volatile:
//
//	ALTER TABLE t ADD COLUMN value int DEFAULT random()::int;      -- volatile
//	ALTER TABLE t ADD COLUMN created_at timestamptz DEFAULT now(); -- stable
//
// Walks the AST so casts do not hide function calls.
// Unknown functions and unsupported expressions are treated as volatile.
func extractDefaultVolatility(constraints []*pgquery.Node) bool {
	for _, c := range constraints {
		constraint := c.GetConstraint()
		if constraint == nil || constraint.Contype != pgquery.ConstrType_CONSTR_DEFAULT {
			continue
		}
		return volatileExpr(constraint.RawExpr)
	}
	return false
}

// volatileExpr walks the expression the way deparseDef does, and treats a node
// kind it does not recognise as volatile: an expression we cannot read through
// is not one we can call cheap.
func volatileExpr(node *pgquery.Node) bool {
	if node == nil {
		return false
	}

	switch {
	case node.GetAConst() != nil, node.GetSqlvalueFunction() != nil, node.GetColumnRef() != nil:
		return false
	case node.GetFuncCall() != nil:
		return volatileFuncCall(node.GetFuncCall())
	case node.GetTypeCast() != nil:
		return volatileExpr(node.GetTypeCast().Arg)
	case node.GetAExpr() != nil:
		return volatileExpr(node.GetAExpr().Lexpr) || volatileExpr(node.GetAExpr().Rexpr)
	case node.GetNullTest() != nil:
		return volatileExpr(node.GetNullTest().Arg)
	case node.GetBoolExpr() != nil:
		return anyVolatile(node.GetBoolExpr().Args)
	case node.GetCoalesceExpr() != nil:
		return anyVolatile(node.GetCoalesceExpr().Args)
	case node.GetMinMaxExpr() != nil:
		return anyVolatile(node.GetMinMaxExpr().Args)
	case node.GetRowExpr() != nil:
		return anyVolatile(node.GetRowExpr().Args)
	case node.GetList() != nil:
		return anyVolatile(node.GetList().Items)
	default:
		return true
	}
}

func volatileFuncCall(fc *pgquery.FuncCall) bool {
	if nonVolatileFunctions[funcName(fc)] {
		return anyVolatile(fc.Args)
	}
	return true
}

// funcName is the unqualified, lowercased name of the called function, so
// pg_catalog.now() and NOW() read the same.
func funcName(fc *pgquery.FuncCall) string {
	if fc == nil || len(fc.Funcname) == 0 {
		return ""
	}
	last := fc.Funcname[len(fc.Funcname)-1].GetString_()
	if last == nil {
		return ""
	}
	return strings.ToLower(last.Sval)
}

func anyVolatile(nodes []*pgquery.Node) bool {
	return slices.ContainsFunc(nodes, volatileExpr)
}
