package kb

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadNoteFileAfterSymlinkSwap(t *testing.T) {
	root := t.TempDir()
	victim := filepath.Join(root, "victim.md")
	outside := filepath.Join(t.TempDir(), "private.md")
	for name, body := range map[string]string{victim: "# Safe", outside: "# OUTSIDE_SECRET_SENTINEL"} {
		if err := os.WriteFile(name, []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}
	dir, err := os.OpenRoot(root)
	if err != nil {
		t.Fatal(err)
	}
	defer dir.Close()
	if _, err := os.ReadDir(root); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(victim); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, victim); err != nil {
		t.Fatal(err)
	}
	raw, _, err := readNoteFile(dir, "victim.md")
	if err == nil || len(raw) != 0 {
		t.Fatalf("read swapped symlink: %q %v", raw, err)
	}
}
