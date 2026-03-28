package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type sinceResult struct {
	gitArg string // value passed to git --since=
	label  string // displayed in the output header
}

var shortRe = regexp.MustCompile(`^(\d+)(h|d|w)$`)

var dayNames = []string{
	"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday",
}

// parseSince converts a natural language time expression into a git --since
// argument and a human-readable label. Git itself understands most of these
// (it uses approxidate internally), so we mostly pass through and just normalise
// a few shorthand forms.
//
// Supported inputs:
//
//	24h, 48h, ...        → "N hours ago"
//	1d, 2d, 7d, ...      → "N days ago"
//	1w, 2w, ...          → "N weeks ago"
//	yesterday            → "yesterday"
//	today                → "midnight" (git: since start of today)
//	week                 → "7 days ago"
//	monday … sunday      → passed through to git (last occurrence)
//	last monday … sunday → passed through to git
//	anything else        → passed through as-is (e.g. "2024-01-15", "3 months ago")
func parseSince(s string) sinceResult {
	s = strings.TrimSpace(strings.ToLower(s))

	// Shorthand: 2d, 3w, 12h
	if m := shortRe.FindStringSubmatch(s); m != nil {
		n, _ := strconv.Atoi(m[1])
		switch m[2] {
		case "h":
			unit := pluralise("hour", n)
			return sinceResult{
				gitArg: fmt.Sprintf("%d %s ago", n, unit),
				label:  fmt.Sprintf("last %d%s", n, m[2]),
			}
		case "d":
			unit := pluralise("day", n)
			return sinceResult{
				gitArg: fmt.Sprintf("%d %s ago", n, unit),
				label:  fmt.Sprintf("last %d%s", n, m[2]),
			}
		case "w":
			unit := pluralise("week", n)
			return sinceResult{
				gitArg: fmt.Sprintf("%d %s ago", n, unit),
				label:  fmt.Sprintf("last %d%s", n, m[2]),
			}
		}
	}

	switch s {
	case "24h", "":
		return sinceResult{gitArg: "24 hours ago", label: "last 24h"}
	case "week", "7d":
		return sinceResult{gitArg: "7 days ago", label: "last 7 days"}
	case "yesterday":
		return sinceResult{gitArg: "yesterday", label: "yesterday"}
	case "today":
		return sinceResult{gitArg: "midnight", label: "today"}
	}

	// Day names and "last <day>" — git handles these natively
	for _, day := range dayNames {
		if s == day || s == "last "+day {
			cap := strings.ToUpper(s[:1]) + s[1:]
			return sinceResult{gitArg: s, label: "since " + cap}
		}
	}

	// Pass-through: absolute dates, month names, etc.
	return sinceResult{gitArg: s, label: "since " + s}
}

func pluralise(word string, n int) string {
	if n == 1 {
		return word
	}
	return word + "s"
}
