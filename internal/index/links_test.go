package index

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestLinkGraphTransitions(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	root := t.TempDir()
	idx, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	defer idx.Close()
	ctx := context.Background()
	write := func(id, body string) {
		t.Helper()
		path := filepath.Join(root, id+".md")
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}
	sync := func(full bool) {
		t.Helper()
		if err := idx.Sync(ctx, root, full); err != nil {
			t.Fatal(err)
		}
	}
	check := func(wantID, wantReason string) {
		t.Helper()
		sync(false)
		links, err := idx.Links(ctx, "source")
		if err != nil || len(links.Links) != 1 {
			t.Fatalf("%+v %v", links, err)
		}
		l := links.Links[0]
		if wantID != "" {
			if l.TargetID == nil || *l.TargetID != wantID {
				t.Fatalf("%+v", l)
			}
		} else {
			broken, err := idx.Broken(ctx)
			if err != nil || broken.Total != 1 || broken.Broken[0].Reason != wantReason {
				t.Fatalf("%+v %v", broken, err)
			}
		}
	}
	write("source", "[[shared]]\n")
	check("", "missing")
	write("a/shared", "---\naliases: [Other]\n---\n")
	check("a/shared", "")
	write("b/shared", "# Second\n")
	check("", "ambiguous")
	if err := os.Remove(filepath.Join(root, "b/shared.md")); err != nil {
		t.Fatal(err)
	}
	check("a/shared", "")
	write("source", "![[Other#Heading|display]]\n")
	check("a/shared", "")
	write("a/shared", "---\naliases: [Renamed]\n---\n")
	check("", "missing")
	write("b/target", "---\naliases: [OTHER]\n---\n")
	check("b/target", "")
	sync(true)
	check("b/target", "")
	if err := os.Remove(filepath.Join(root, "source.md")); err != nil {
		t.Fatal(err)
	}
	sync(false)
	links, broken, _, err := idx.LinkCounts(ctx)
	if err != nil || links != 0 || broken != 0 {
		t.Fatalf("%d %d %v", links, broken, err)
	}
}

func TestUpgradeLinksInUnchangedM1Cache(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "source.md"), []byte("[[missing]]"), 0644); err != nil {
		t.Fatal(err)
	}
	idx, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := idx.Sync(ctx, root, false); err != nil {
		t.Fatal(err)
	}
	if _, err := idx.db.Exec("DROP TABLE links; DROP TABLE link_sources"); err != nil {
		t.Fatal(err)
	}
	idx.Close()
	idx, err = Open(root)
	if err != nil {
		t.Fatal(err)
	}
	defer idx.Close()
	if err := idx.Sync(ctx, root, false); err != nil {
		t.Fatal(err)
	}
	result, err := idx.Links(ctx, "source")
	if err != nil || len(result.Links) != 1 {
		t.Fatalf("%+v %v", result, err)
	}
}
