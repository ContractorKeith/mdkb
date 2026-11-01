package kb

import (
	"reflect"
	"testing"
)

func TestExtractLinks(t *testing.T) {
	for _, tt := range []struct {
		name, body string
		want       []LinkRef
	}{
		{"forms", "[[ plain ]] [[plain|display]] [[plain#Section]] [[plain#Section|display]] ![[plain]] ![[plain|display]] ![[plain#Section]] ![[plain#Section|display]]", []LinkRef{
			{" plain ", "plain", "", "", false, 1}, {"plain|display", "plain", "", "display", false, 1},
			{"plain#Section", "plain", "Section", "", false, 1}, {"plain#Section|display", "plain", "Section", "display", false, 1},
			{"plain", "plain", "", "", true, 1}, {"plain|display", "plain", "", "display", true, 1},
			{"plain#Section", "plain", "Section", "", true, 1}, {"plain#Section|display", "plain", "Section", "display", true, 1},
		}},
		{"fences", "```md\n[[no]]\n~~~\n[[no]]\n```\n~~~md\n[[no]]\n~~~\n[[yes]]", []LinkRef{{"yes", "yes", "", "", false, 9}}},
		{"inline", "`[[no]]` ``a ` ![[no]]`` [[yes]]", []LinkRef{{"yes", "yes", "", "", false, 1}}},
		{"multiline inline", "`start\n[[no]]`\n[[yes]]", []LinkRef{{"yes", "yes", "", "", false, 3}}},
		{"unmatched tick", "` [[yes]]", []LinkRef{{"yes", "yes", "", "", false, 1}}},
		{"empty", "[[]] [[ |alias]] [[#heading]] [[unfinished", []LinkRef{}},
		{"lines", "first\n![[a]] and [[b]]\n", []LinkRef{{"a", "a", "", "", true, 2}, {"b", "b", "", "", false, 2}}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExtractLinks(tt.body); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestResolve(t *testing.T) {
	notes := []Note{{ID: "plain", Aliases: []string{"Root Alias", "Document.md"}}, {ID: "nested/unique", Aliases: []string{"Name", "Name"}}, {ID: "a/shared", Aliases: []string{"Duplicate"}}, {ID: "b/shared", Aliases: []string{"Duplicate"}}, {ID: "shadow", Aliases: []string{"shared", "plain"}}}
	for _, tt := range []struct{ target, id, reason string }{
		{"plain", "plain", ""}, {"plain.md", "plain", ""}, {"a/shared.md", "a/shared", ""},
		{"unique", "nested/unique", ""}, {"shared", "", "ambiguous"},
		{"Name", "nested/unique", ""}, {"nAmE", "nested/unique", ""},
		{"Root Alias", "plain", ""}, {"DOCUMENT.md", "plain", ""}, {"Duplicate", "", "ambiguous"},
		{"missing", "", "missing"}, {"missing/unique", "", "missing"},
	} {
		t.Run(tt.target, func(t *testing.T) {
			id, reason := ResolveLink(tt.target, notes)
			if id != tt.id || reason != tt.reason {
				t.Fatalf("%q %q", id, reason)
			}
			id, ok := Resolve(tt.target, notes)
			if id != tt.id || ok != (tt.reason == "") {
				t.Fatalf("%q %v", id, ok)
			}
		})
	}
}
