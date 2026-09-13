package models

import (
	"encoding/json"
	"strconv"
	"strings"
)

// SQLType is a declared type as the parser read it out of the AST: the base name
// and the modifiers it was written with, e.g. varchar(255) is {"varchar", [255]}.
// pg_query already canonicalises the spelling, so CHARACTER VARYING and DECIMAL
// arrive as varchar and numeric.
//
// The parts are kept apart because classification compares them — whether a type
// change rewrites the table depends on the base names and on whether the
// modifiers widen. Rendering to text is a one-way step taken at the API
// boundary, so nothing downstream has to take the text apart again.
type SQLType struct {
	Base string
	Mods []int
}

func NewSQLType(base string, mods ...int) SQLType {
	return SQLType{Base: base, Mods: mods}
}

func (t SQLType) IsZero() bool {
	return t.Base == ""
}

func (t SQLType) String() string {
	if len(t.Mods) == 0 {
		return t.Base
	}
	mods := make([]string, 0, len(t.Mods))
	for _, mod := range t.Mods {
		mods = append(mods, strconv.Itoa(mod))
	}
	return t.Base + "(" + strings.Join(mods, ",") + ")"
}

// The API carries the declared type as the text a migration would write.
func (t SQLType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}
