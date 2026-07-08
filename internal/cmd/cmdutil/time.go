package cmdutil

import (
	"fmt"
	"time"
)

// RelativeTime formats t relative to now, e.g. "3 days ago", "just now".
func RelativeTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return pluralAgo(int(d.Minutes()), "minute")
	case d < 24*time.Hour:
		return pluralAgo(int(d.Hours()), "hour")
	case d < 30*24*time.Hour:
		return pluralAgo(int(d.Hours()/24), "day")
	case d < 365*24*time.Hour:
		return pluralAgo(int(d.Hours()/(24*30)), "month")
	default:
		return pluralAgo(int(d.Hours()/(24*365)), "year")
	}
}

func pluralAgo(n int, unit string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s ago", n, unit)
	}
	return fmt.Sprintf("%d %ss ago", n, unit)
}
