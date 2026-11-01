package kb

import (
	"path"
	"strings"
)

// Resolve tries exact ID, unique basename, then unique case-insensitive alias.
func Resolve(target string, notes []Note) (string, bool) {
	id, reason := ResolveLink(target, notes)
	return id, reason == ""
}

// ResolveLink also explains unresolved references for the broken-link report.
func ResolveLink(target string, notes []Note) (id, reason string) {
	target = strings.TrimSpace(target)
	for _, n := range notes {
		if n.ID == target {
			return n.ID, ""
		}
	}
	aliasTarget := target
	target = strings.TrimSuffix(target, ".md")
	for _, n := range notes {
		if n.ID == target {
			return n.ID, ""
		}
	}
	if !strings.Contains(target, "/") {
		count := 0
		for _, n := range notes {
			if path.Base(n.ID) == target {
				id = n.ID
				count++
			}
		}
		if count == 1 {
			return id, ""
		}
		if count > 1 {
			return "", "ambiguous"
		}
	}
	count := 0
	for _, n := range notes {
		for _, alias := range n.Aliases {
			if strings.EqualFold(alias, aliasTarget) {
				id = n.ID
				count++
				break
			}
		}
	}
	if count == 1 {
		return id, ""
	}
	if count > 1 {
		return "", "ambiguous"
	}
	return "", "missing"
}
