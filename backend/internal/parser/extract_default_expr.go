package parser

import pgquery "github.com/pganalyze/pg_query_go/v6"

// extractDefaultExpr deparses the column's DEFAULT expression, returning "" when
// the column has no default or the expression is not one deparseDef handles.
func extractDefaultExpr(constraints []*pgquery.Node) string {
	for _, c := range constraints {
		constraint := c.GetConstraint()
		if constraint == nil || constraint.Contype != pgquery.ConstrType_CONSTR_DEFAULT {
			continue
		}
		return deparseDef(constraint.RawExpr)
	}
	return ""
}
