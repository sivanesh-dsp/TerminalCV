package shell

import "strings"

// Tokenize splits a command line into argv-style tokens. It honours single
// quotes (literal), double quotes (grouping) and backslash escapes, and
// collapses runs of unquoted whitespace — like a real shell word-splitter.
func Tokenize(line string) []string {
	var tokens []string
	var cur strings.Builder
	inSingle, inDouble, escaped, hasToken := false, false, false, false

	flush := func() {
		if hasToken {
			tokens = append(tokens, cur.String())
			cur.Reset()
			hasToken = false
		}
	}

	for _, r := range line {
		switch {
		case escaped:
			cur.WriteRune(r)
			hasToken = true
			escaped = false
		case r == '\\' && !inSingle:
			// Backslash escapes the next rune (except inside single quotes).
			escaped = true
			hasToken = true
		case r == '\'' && !inDouble:
			inSingle = !inSingle
			hasToken = true
		case r == '"' && !inSingle:
			inDouble = !inDouble
			hasToken = true
		case (r == ' ' || r == '\t') && !inSingle && !inDouble:
			flush()
		default:
			cur.WriteRune(r)
			hasToken = true
		}
	}
	flush()
	return tokens
}
