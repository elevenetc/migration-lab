package runtime

import (
	"fmt"
	"strings"
)

// SeedStatement builds the INSERT that fills the planned table from
// generate_series. It returns "" when no column of the table can be given a
// value, since there is then nothing to insert.
func SeedStatement(plan SeedPlan, rows int64) string {
	if len(plan.Columns) == 0 || rows <= 0 {
		return ""
	}

	names := make([]string, len(plan.Columns))
	exprs := make([]string, len(plan.Columns))
	for i, column := range plan.Columns {
		names[i] = quoteIdentifier(column.Name)
		exprs[i] = column.Expr
	}

	return fmt.Sprintf("INSERT INTO %s (%s) SELECT %s FROM generate_series(1, %d) AS g(i)",
		quoteIdentifier(plan.Table), strings.Join(names, ", "), strings.Join(exprs, ", "), rows)
}

// quoteIdentifier is shared by every statement this package builds by hand.
func quoteIdentifier(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}
