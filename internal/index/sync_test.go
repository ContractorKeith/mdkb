package index

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestIncrementalSync(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	cache := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", cache)
	path := filepath.Join(root, "note.md")
	write := func(body string) {
		t.Helper()
		if err := os.WriteFile(path, []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}
	write("---\ntags: [one, one]\n---\n# First\nkeyword\n")
	idx, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	defer idx.Close()
	hash := sha256.Sum256([]byte(root))
	if idx.Path != filepath.Join(cache, "mdkb", fmt.Sprintf("%x.db", hash[:8])) {
		t.Fatal(idx.Path)
	}
	sync := func(full bool) {
		t.Helper()
		if err := idx.Sync(ctx, root, full); err != nil {
			t.Fatal(err)
		}
	}
	sync(false)
	// Unchanged stamps must not rewrite the stored row.
	if _, err := idx.db.Exec(`UPDATE notes SET title='Sentinel'`); err != nil {
		t.Fatal(err)
	}
	sync(false)
	listing, err := idx.List(ctx, "one", 50)
	if err != nil || listing.Total != 1 || listing.Notes[0].Title != "Sentinel" {
		t.Fatalf("%+v %v", listing, err)
	}
	tags, err := idx.Tags(ctx)
	if err != nil || len(tags) != 1 || tags[0].Count != 1 {
		t.Fatalf("%+v %v", tags, err)
	}
	write("---\ntags: [two, two]\n---\n# Other\nnewword\n")
	future := time.Now().Add(time.Second)
	if err := os.Chtimes(path, future, future); err != nil {
		t.Fatal(err)
	}
	sync(false)
	hits, err := idx.Keyword(ctx, "keyword", 50)
	if err != nil || len(hits) != 0 {
		t.Fatalf("stale FTS: %+v %v", hits, err)
	}
	hits, err = idx.Keyword(ctx, "newword", 50)
	if err != nil || len(hits) != 1 {
		t.Fatalf("updated FTS: %+v %v", hits, err)
	}
	sync(true)
	hits, err = idx.Keyword(ctx, "newword", 50)
	if err != nil || len(hits) != 1 {
		t.Fatalf("rebuilt FTS: %+v %v", hits, err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	sync(false)
	hits, err = idx.Keyword(ctx, "newword", 50)
	if err != nil || len(hits) != 0 {
		t.Fatalf("deleted FTS: %+v %v", hits, err)
	}
	if err := idx.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := idx.Keyword(ctx, "x", 50); !errors.Is(err, ErrIndex) {
		t.Fatalf("closed index: %v", err)
	}
}

func TestCacheFallback(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CACHE_HOME", "")
	idx, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer idx.Close()
	if filepath.Dir(idx.Path) != filepath.Join(home, ".cache", "mdkb") {
		t.Fatal(idx.Path)
	}
}

func TestUpgradeCacheProblemColumn(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	root := t.TempDir()
	idx, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := idx.db.Exec("ALTER TABLE notes DROP COLUMN problem"); err != nil {
		t.Fatal(err)
	}
	if err := idx.Close(); err != nil {
		t.Fatal(err)
	}
	idx, err = Open(root)
	if err != nil {
		t.Fatal(err)
	}
	defer idx.Close()
	if err := os.WriteFile(filepath.Join(root, "bad.md"), []byte("---\ntitle: unfinished\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := idx.Sync(context.Background(), root, false); err != nil {
		t.Fatal(err)
	}
	listing, err := idx.List(context.Background(), "", 50)
	if err != nil || len(listing.Notes) != 1 || listing.Notes[0].Problem == "" {
		t.Fatalf("upgraded listing: %+v %v", listing, err)
	}
}
