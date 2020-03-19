package main

import (
	"regexp"
	"time"
)

var floatRegex = regexp.MustCompile(`\.\d+`)

// formatDuration returns a string-formatted Duration without floating points.
func formatDuration(d time.Duration) string {
	orig := d.String()
	return floatRegex.ReplaceAllLiteralString(orig, "")
}
