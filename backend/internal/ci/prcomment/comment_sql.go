package prcomment

import "strings"

// Keep measured SQL literal, including indentation and quoted strings. A fence
// longer than any embedded backtick run keeps repository text inside the block.
func commentSQL(sql string) string {
	runes := []rune(sql)
	truncated := len(runes) > 1000
	if truncated {
		sql = string(runes[:1000])
	}
	fence := "```"
	for strings.Contains(sql, fence) {
		fence += "`"
	}
	body := "\n  " + fence + "sql\n  " + strings.ReplaceAll(sql, "\n", "\n  ") + "\n  " + fence + "\n"
	if truncated {
		body += "\n  SQL truncated; see the workflow artifacts for the full statement.\n"
	}
	return body
}
