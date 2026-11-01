package kb_test

import (
	"bytes"
	"github.com/ContractorKeith/mdkb/internal/kb"
	"os"
	"path/filepath"
	"testing"
)

func TestWalkConfinedReads(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "private.md")
	write(t, outside, "# OUTSIDE_SECRET_SENTINEL\n")
	write(t, filepath.Join(root, "safe.md"), "# Safe\n")
	if err := os.Symlink(outside, filepath.Join(root, "victim.md")); err != nil {
		t.Fatal(err)
	}
	notes, err := kb.Walk(root)
	if err != nil || len(notes) != 1 || notes[0].ID != "safe" {
		t.Fatalf("walk: %+v %v", notes, err)
	}
	info, err := kb.Info(root)
	if err != nil || info.NoteCount != 1 {
		t.Fatalf("info: %+v %v", info, err)
	}
	// Guard the directory-listing/open race: file reads must use the pinned root.
	for _, file := range []string{"walk.go", "kb.go"} {
		raw, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Contains(raw, []byte("os.ReadFile(")) {
			t.Errorf("%s uses unconfined os.ReadFile after directory listing", file)
		}
	}
}
