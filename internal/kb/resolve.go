package kb

import (
	"path"
	"strings"
	"unicode"
)

// Resolver indexes exact IDs, basenames and aliases for one graph snapshot.
type Resolver struct {
	ids       map[string]bool
	basenames map[string]string
	aliases   map[string]string
}

// NewResolver records ambiguous matches as empty IDs.
func NewResolver(notes []Note) *Resolver {
	r := &Resolver{map[string]bool{}, map[string]string{}, map[string]string{}}
	add := func(m map[string]string, key, id string) {
		if old, exists := m[key]; exists && old != id {
			m[key] = ""
		} else {
			m[key] = id
		}
	}
	for _, n := range notes {
		n.ID = CanonicalID(n.ID)
		r.ids[n.ID] = true
		add(r.basenames, path.Base(n.ID), n.ID)
		for _, alias := range n.Aliases {
			add(r.aliases, foldAlias(alias), n.ID)
		}
	}
	return r
}

// foldAlias preserves strings.EqualFold semantics, including Unicode fold cycles.
func foldAlias(s string) string {
	return strings.Map(func(r rune) rune {
		smallest := r
		for next := unicode.SimpleFold(r); next != r; next = unicode.SimpleFold(next) {
			if next < smallest {
				smallest = next
			}
		}
		return smallest
	}, CanonicalID(s))
}

// ResolveLink tries exact ID, unique basename, then unique case-insensitive alias.
func (r *Resolver) ResolveLink(target string) (id, reason string) {
	target = CanonicalID(strings.TrimSpace(target))
	if r.ids[target] {
		return target, ""
	}
	aliasTarget := target
	target = strings.TrimSuffix(target, ".md")
	if r.ids[target] {
		return target, ""
	}
	if !strings.Contains(target, "/") {
		if id, exists := r.basenames[target]; exists {
			if id == "" {
				return "", "ambiguous"
			}
			return id, ""
		}
	}
	if id, exists := r.aliases[foldAlias(aliasTarget)]; exists {
		if id == "" {
			return "", "ambiguous"
		}
		return id, ""
	}
	return "", "missing"
}

// Resolve tries exact ID, unique basename, then unique case-insensitive alias.
func Resolve(target string, notes []Note) (string, bool) {
	id, reason := ResolveLink(target, notes)
	return id, reason == ""
}

// ResolveLink also explains unresolved references for the broken-link report.
func ResolveLink(target string, notes []Note) (id, reason string) {
	return NewResolver(notes).ResolveLink(target)
}
