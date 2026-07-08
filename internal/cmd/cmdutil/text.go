package cmdutil

// Truncate shortens s to at most max runes, appending an ellipsis when
// truncated, for narrow table columns.
func Truncate(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	if max <= 1 {
		return string(r[:max])
	}
	return string(r[:max-1]) + "…"
}
