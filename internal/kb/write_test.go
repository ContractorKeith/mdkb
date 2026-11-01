package kb_test

import (
	"github.com/ContractorKeith/mdkb/internal/kb"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestWriteSafety(t *testing.T) {
	root := t.TempDir()
	if err := kb.Init(root); err != nil {
		t.Fatal(err)
	}
	outside := t.TempDir()
	victim := filepath.Join(outside, "victim.md")
	if err := os.WriteFile(victim, []byte("untouched"), 0644); err != nil {
		t.Fatal(err)
	}
	for link, target := range map[string]string{"file.md": victim, "dir": outside} {
		if err := os.Symlink(target, filepath.Join(root, link)); err != nil {
			t.Fatal(err)
		}
	}
	for _, mode := range []string{"create", "append", "overwrite"} {
		for _, id := range []string{"../escape", ".hidden/x", "a/./b", "file", "dir/victim", "dir/new", "/absolute", "a//b", ".note", "note.md"} {
			if _, err := kb.Write(root, id, "bad", mode); err == nil {
				t.Fatalf("accepted %q %s", id, mode)
			}
		}
	}
	raw, err := os.ReadFile(victim)
	if err != nil || string(raw) != "untouched" {
		t.Fatalf("%q %v", raw, err)
	}
	if _, err := os.Stat(filepath.Join(outside, "new.md")); !os.IsNotExist(err) {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".tmp") {
			t.Fatal("temp leaked")
		}
	}
}

func TestConcurrentCreateAndAppend(t *testing.T) {
	root := t.TempDir()
	if err := kb.Init(root); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	results := make(chan error, 10)
	for i := 0; i < 10; i++ {
		wg.Go(func() { _, err := kb.Write(root, "race", "start", "create"); results <- err })
	}
	wg.Wait()
	close(results)
	successes := 0
	for err := range results {
		if err == nil {
			successes++
		}
	}
	if successes != 1 {
		t.Fatal(successes)
	}
	for i := 0; i < 10; i++ {
		wg.Go(func() {
			if _, err := kb.Write(root, "race", "appended", "append"); err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	_, raw, err := kb.Read(root, "race")
	if err != nil || strings.Count(raw, "appended") != 10 {
		t.Fatalf("%q %v", raw, err)
	}
}
