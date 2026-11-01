package kb_test

import (
	"github.com/ContractorKeith/mdkb/internal/kb"
	"os"
	"path/filepath"
	"testing"
)

func TestListedIDsReadable(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"a..b.md", "ending.md.md", ".md", ".secret.md"} {
		write(t, filepath.Join(root, name), "# Note\n")
	}
	notes, err := kb.Walk(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(notes) != 2 {
		t.Fatalf("listed %d notes, want 2", len(notes))
	}
	for _, n := range notes {
		if _, _, err := kb.Read(root, n.ID); err != nil {
			t.Errorf("listed id %q unreadable: %v", n.ID, err)
		}
	}
	info, err := kb.Info(root)
	if err != nil || info.NoteCount != 2 || info.Problems != 2 {
		t.Fatalf("info: %+v %v", info, err)
	}
	if err := kb.ValidateID("ending.md", true); err == nil {
		t.Error("write accepted .md suffix")
	}
	if _, err := os.Stat(filepath.Join(root, "ending.md.md")); err != nil {
		t.Fatal(err)
	}
}
