package runtime

import (
	"fmt"
	"strings"
)

// SeedExpr is the value expression for a column of the given type, or "" when
// the type has no generic filler and the column is left to its default.
// maxLength is the declared length of a character type, 0 when unbounded.
func SeedExpr(dataType string, maxLength int) string {
	switch strings.ToLower(dataType) {
	case "smallint", "integer", "bigint", "numeric", "decimal", "real", "double precision":
		return "g.i"
	case "boolean":
		return "(g.i % 2 = 0)"
	case "text", "character varying", "character":
		return truncatedText(maxLength)
	case "uuid":
		return "gen_random_uuid()"
	case "date":
		return "current_date - (g.i % 3650)"
	case "timestamp without time zone", "timestamp with time zone":
		return "now() - (g.i % 86400) * interval '1 second'"
	case "json":
		return "json_build_object('i', g.i)"
	case "jsonb":
		return "jsonb_build_object('i', g.i)"
	case "bytea":
		return "int4send(g.i)"
	default:
		return ""
	}
}

// truncatedText keeps generated text inside a declared length, so a narrow
// varchar(n) is seeded instead of overflowing it.
func truncatedText(maxLength int) string {
	const expr = "('row_' || g.i)"
	if maxLength <= 0 {
		return expr
	}
	return fmt.Sprintf("left(%s, %d)", expr, maxLength)
}
