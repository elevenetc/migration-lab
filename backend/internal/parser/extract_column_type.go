package parser

import (
	"strings"

	pgquery "github.com/pganalyze/pg_query_go/v6"
	"migration-timeline/backend/internal/models"
)

// extractColumnType reads a declared type out of the AST into its parts. The
// pg_catalog qualifier is dropped, since pg_query adds it to every built-in type
// and no migration writes it.
//
// Only integer modifiers are kept: they are the ones that decide whether a type
// change rewrites the table, and every modifier PostgreSQL accepts on a
// rewrite-relevant type — varchar, bpchar, numeric, timestamp precision — is one.
func extractColumnType(typeName *pgquery.TypeName) models.SQLType {
	if typeName == nil {
		return models.SQLType{}
	}

	var parts []string
	for _, name := range typeName.Names {
		str := name.GetString_()
		if str != nil && str.Sval != "pg_catalog" {
			parts = append(parts, str.Sval)
		}
	}

	var mods []int
	for _, mod := range typeName.Typmods {
		constVal := mod.GetAConst()
		if constVal == nil {
			continue
		}
		if ival := constVal.GetIval(); ival != nil {
			mods = append(mods, int(ival.Ival))
		}
	}

	return models.SQLType{Base: strings.Join(parts, "."), Mods: mods}
}
