package kb

import "strings"

// fenceLine rejects backticks in a backtick opener's info string.
func fenceLine(line string) bool {
	if strings.HasPrefix(line, "~~~") {
		return true
	}
	if !strings.HasPrefix(line, "```") {
		return false
	}
	return !strings.Contains(strings.TrimLeft(line, "`"), "`")
}
