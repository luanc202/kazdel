package utils

import (
	"fmt"
	"time"
)

func RelativeTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)
	isFuture := diff < 0
	if isFuture {
		diff = -diff
	}

	var res string
	seconds := int(diff.Seconds())
	minutes := int(diff.Minutes())
	hours := int(diff.Hours())
	days := int(hours / 24)
	months := int(days / 30)
	years := int(days / 365)

	switch {
	case seconds < 60:
		res = fmt.Sprintf("%d seconds", seconds)
		if seconds == 1 {
			res = "1 second"
		}
	case minutes < 60:
		res = fmt.Sprintf("%d minutes", minutes)
		if minutes == 1 {
			res = "1 minute"
		}
	case hours < 24:
		res = fmt.Sprintf("%d hours", hours)
		if hours == 1 {
			res = "1 hour"
		}
	case days < 30:
		res = fmt.Sprintf("%d days", days)
		if days == 1 {
			res = "1 day"
		}
	case months < 12:
		res = fmt.Sprintf("%d months", months)
		if months == 1 {
			res = "1 month"
		}
	default:
		res = fmt.Sprintf("%d years", years)
		if years == 1 {
			res = "1 year"
		}
	}

	if isFuture {
		return "in " + res
	}
	return res + " ago"
}

func ExpirationTime(t time.Time) string {
	if t.IsZero() {
		return "never"
	}
	now := time.Now()
	if t.Before(now) {
		return "expired"
	}
	return RelativeTime(t)
}
