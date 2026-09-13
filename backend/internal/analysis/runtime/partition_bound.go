package runtime

import (
	"regexp"
	"slices"
	"strings"
)

// Partition strategies as pg_get_partkeydef spells them.
const (
	StrategyRange = "RANGE"
	StrategyList  = "LIST"
	StrategyHash  = "HASH"
)

var identifier = regexp.MustCompile(`^[a-z_][a-z0-9_$]*$`)

// PartitionStrategy is the strategy of a pg_get_partkeydef result, e.g. RANGE
// for "RANGE (created_at)". Empty when the definition is not recognised.
func PartitionStrategy(partkeydef string) string {
	fields := strings.Fields(partkeydef)
	if len(fields) == 0 {
		return ""
	}
	switch strategy := strings.ToUpper(fields[0]); strategy {
	case StrategyRange, StrategyList, StrategyHash:
		return strategy
	default:
		return ""
	}
}

// PartitionKeyColumns are the columns a pg_get_partkeydef result partitions on.
// It returns nil for an expression key, which no seed value can be derived for.
func PartitionKeyColumns(partkeydef string) []string {
	group, ok := firstGroup(partkeydef)
	if !ok {
		return nil
	}

	columns := splitList(group)
	for i, column := range columns {
		name, ok := plainIdentifier(column)
		if !ok {
			return nil
		}
		columns[i] = name
	}
	return columns
}

// plainIdentifier is the column a partition key element names. A quoted
// identifier keeps whatever case it was declared with; anything that is not a
// bare or quoted name is an expression key, which no seed value follows from.
func plainIdentifier(token string) (string, bool) {
	if len(token) > 2 && strings.HasPrefix(token, `"`) && strings.HasSuffix(token, `"`) {
		return strings.ReplaceAll(token[1:len(token)-1], `""`, `"`), true
	}
	return token, identifier.MatchString(token)
}

// PartitionSeedValues are literals that land inside a leaf partition's bound, as
// pg_get_expr(relpartbound) spells it: the lower bound of a range, or the first
// admitted value of a list. It returns nil when no constant is derivable — a
// hash or default partition, or a range open at the bottom — so the caller
// seeds the partitioned parent and lets tuple routing place the rows.
func PartitionSeedValues(bound string) []string {
	trimmed := strings.TrimSpace(bound)

	switch {
	case strings.HasPrefix(trimmed, "FOR VALUES FROM"):
		group, ok := firstGroup(trimmed)
		if !ok {
			return nil
		}
		values := splitList(group)
		if slices.ContainsFunc(values, unbounded) {
			return nil
		}
		return values

	case strings.HasPrefix(trimmed, "FOR VALUES IN"):
		group, ok := firstGroup(trimmed)
		if !ok {
			return nil
		}
		values := splitList(group)
		if len(values) == 0 || strings.EqualFold(values[0], "NULL") {
			return nil
		}
		return values[:1]

	default:
		return nil
	}
}

func unbounded(value string) bool {
	return strings.EqualFold(value, "MINVALUE") || strings.EqualFold(value, "MAXVALUE")
}

// firstGroup is the text between the first parenthesis and the one closing it,
// ignoring parentheses inside string literals.
func firstGroup(s string) (string, bool) {
	start := -1
	depth := 0
	quoted := false

	for i, r := range s {
		switch {
		case r == '\'':
			quoted = !quoted
		case quoted:
		case r == '(':
			if depth == 0 {
				start = i + 1
			}
			depth++
		case r == ')':
			depth--
			if depth == 0 && start >= 0 {
				return s[start:i], true
			}
		}
	}
	return "", false
}

// splitList splits a comma-separated literal list, keeping commas inside string
// literals and nested parentheses together.
func splitList(group string) []string {
	var values []string
	depth := 0
	quoted := false
	start := 0

	for i, r := range group {
		switch {
		case r == '\'':
			quoted = !quoted
		case quoted:
		case r == '(':
			depth++
		case r == ')':
			depth--
		case r == ',' && depth == 0:
			values = append(values, strings.TrimSpace(group[start:i]))
			start = i + 1
		}
	}

	if last := strings.TrimSpace(group[start:]); last != "" {
		values = append(values, last)
	}
	return values
}
