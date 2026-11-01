package kb_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ContractorKeith/mdkb/internal/kb"
)

func TestParse(t *testing.T) {
	for _, tt := range []struct {
		raw, title string
		tags       int
	}{
		{"---\ntitle: Explicit\ntags: single\naliases: [Alias]\ncreated: 2026-10-01\nupdated: 2026-11-01\n---\n# Heading\n", "Explicit", 1},
		{"---\ntags: [one, two]\n---\n# Heading\n", "Heading", 2},
		{"no heading\n", "file", 0},
		{"```md\n# Ignored\n```\n# Actual\n", "Actual", 0},
		{"---\n---\n# Empty metadata\n", "Empty metadata", 0},
	} {
		n, err := kb.Parse("nested/file.md", []byte(tt.raw))
		if err != nil {
			t.Fatal(err)
		}
		if n.Title != tt.title || n.ID != "nested/file" || len(n.Tags) != tt.tags || n.Aliases == nil {
			t.Fatalf("%+v", n)
		}
		if tt.title == "Explicit" && (n.Created != "2026-10-01" || n.Updated != "2026-11-01" || len(n.Aliases) != 1) {
			t.Fatalf("metadata: %+v", n)
		}
	}
	for _, raw := range []string{"---\ntitle: [invalid]\n---\n", "---\ntitle: missing end\n"} {
		n, err := kb.Parse("bad.md", []byte(raw))
		if err != nil || n.Problem == "" || n.Body != raw || n.Title != "bad" || len(n.Tags) != 0 || len(n.Aliases) != 0 {
			t.Fatalf("fallback: %+v, %v", n, err)
		}
	}
}
func TestWalk(t *testing.T) {
	root := t.TempDir()
	for _, p := range []string{"a.md", "nested/b.md", ".hidden/c.md", "ignore.txt"} {
		write(t, filepath.Join(root, p), "# Heading\n")
	}
	if err := os.Symlink(filepath.Join(root, "a.md"), filepath.Join(root, "link.md")); err != nil {
		t.Fatal(err)
	}
	notes, err := kb.Walk(root)
	if err != nil {
		t.Fatal(err)
	}
	info, err := kb.Info(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(notes) != info.NoteCount || len(notes) != 2 || notes[0].Size == 0 || notes[0].MTime == 0 {
		t.Fatalf("%+v", notes)
	}
}
