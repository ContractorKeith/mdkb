package kb_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/ContractorKeith/mdkb/internal/kb"
)

func TestReadDirect(t *testing.T) {
	isolateRoundTest(t)
	root := t.TempDir()
	raw := "---\ntitle: Direct\n---\n# Heading\n"
	write(t, filepath.Join(root, "nested", "note.md"), raw)
	// This unrelated unreadable file must not prevent reading the requested note.
	write(t, filepath.Join(root, "unreadable.md"), "ignored")
	if err := os.Chmod(filepath.Join(root, "unreadable.md"), 0); err != nil {
		t.Fatal(err)
	}
	n, got, err := kb.Read(root, "nested/note")
	if err != nil || got != raw || n.ID != "nested/note" || n.Title != "Direct" {
		t.Fatalf("%+v %q %v", n, got, err)
	}
	if err := os.Symlink(filepath.Join(root, "nested", "note.md"), filepath.Join(root, "link.md")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(root, "nested"), filepath.Join(root, "linked-dir")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("nested/note.md", filepath.Join(root, "relative-link.md")); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(root, ".hidden", "note.md"), "# Hidden")
	for _, id := range []string{"nested/./note", "../outside", "nested/../nested/note", filepath.Join(root, "nested/note"), "link", "relative-link", "linked-dir/note", ".hidden/note"} {
		if _, got, err := kb.Read(root, id); err == nil || got != "" {
			t.Fatalf("read %q: %q %v", id, got, err)
		}
	}
	for _, id := range []string{"nope", "missing/child"} {
		_, _, err := kb.Read(root, id)
		if !errors.Is(err, os.ErrNotExist) || err.Error() != "note not found: "+id {
			t.Fatalf("%q: %v", id, err)
		}
	}
	linkRoot := filepath.Join(t.TempDir(), "root")
	if err := os.Symlink(root, linkRoot); err != nil {
		t.Fatal(err)
	}
	if _, got, err := kb.Read(linkRoot, "nested/note"); err != nil || got != raw {
		t.Fatalf("root symlink: %q %v", got, err)
	}
}
