package models

// classifyAlterColumnType decides whether a type change rewrites the table,
// which the target type alone cannot answer: varchar(100) -> text is free while
// text -> varchar(100) rewrites and verifies. PostgreSQL skips the rewrite when
// the cast is binary-coercible and the new type imposes no constraint the old one
// did not, so only widening and constraint-dropping changes are metadata-only.
//
// An unknown previous type — no earlier migration in the timeline declared the
// column — keeps the conservative answer.
func classifyAlterColumnType(op AlterColumnType) PerformanceClass {
	if op.PreviousType.IsZero() {
		return TableRewrite
	}
	if coercibleWithoutRewrite(op.PreviousType, op.NewType) {
		return MetadataOnly
	}
	return TableRewrite
}

func coercibleWithoutRewrite(from, to SQLType) bool {
	if from.Base == to.Base && sameMods(from.Mods, to.Mods) {
		return true
	}

	switch {
	// A varchar keeps its representation as text and as a longer varchar; only a
	// shorter limit is a new constraint to verify.
	case from.Base == "varchar" && to.Base == "text":
		return true
	case from.Base == "varchar" && to.Base == "varchar":
		return dropsOrWidens(from.Mods, to.Mods)
	// numeric stores its own precision, so widening it checks nothing. The scale
	// has to stay, since changing it rounds every value.
	case from.Base == "numeric" && to.Base == "numeric":
		return dropsOrWidens(from.Mods, to.Mods)
	default:
		return false
	}
}

// dropsOrWidens reports whether the new modifiers admit everything the old ones
// did: no modifier at all admits every value, and otherwise every modifier but
// the first — the scale of a numeric — has to stay as it was.
func dropsOrWidens(from, to []int) bool {
	if len(to) == 0 {
		return true
	}
	if len(from) == 0 || len(from) != len(to) {
		return false
	}
	if !sameMods(from[1:], to[1:]) {
		return false
	}
	return to[0] >= from[0]
}

func sameMods(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
