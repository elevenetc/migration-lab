package runtime

import (
	"fmt"
	"strings"
)

// SeedStatement builds the INSERT that fills the planned table from
// generate_series. A plan with no columns — every column filled by the
// database itself — inserts zero-column rows, leaving each column to its
// default.
func SeedStatement(plan SeedPlan, rows int64) string {
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

	return fmt.Sprintf("INSERT INTO %s SELECT %sFROM generate_series(1, %d) AS g(i)", target, values, rows)
}

// quoteIdentifier is shared by every statement this package builds by hand.
func quoteIdentifier(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}
