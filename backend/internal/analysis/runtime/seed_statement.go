package runtime

import (
	"fmt"
	"strings"
)

// SeedStatement builds the INSERT that fills the planned table with rows from
// through to of generate_series; starting past the rows a table already holds
// keeps generated keys from colliding with them. A plan with no columns — every
// column filled by the database itself — inserts zero-column rows, leaving each
// column to its default.
func SeedStatement(plan SeedPlan, from, to int64) string {
	names := make([]string, len(plan.Columns))
	exprs := make([]string, len(plan.Columns))
	for i, column := range plan.Columns {
		names[i] = quoteIdentifier(column.Name)
		exprs[i] = column.Expr
	}

	target, values := quoteIdentifier(plan.Table), ""
	if len(names) > 0 {
		target += " (" + strings.Join(names, ", ") + ")"
		values = strings.Join(exprs, ", ") + " "
	}

	return fmt.Sprintf("INSERT INTO %s SELECT %sFROM generate_series(%d, %d) AS g(i)", target, values, from, to)
}

// quoteIdentifier is shared by every statement this package builds by hand.
func quoteIdentifier(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}
