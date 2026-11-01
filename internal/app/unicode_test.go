package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestUnicodeGraphAndMove(t *testing.T) {
	root := writeFixture(t)
	for name, body := range map[string]string{"Cafe\u0301.md": "# Coffee\n", "source.md": "[[Café]] [[Cafe\u0301]]\n"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}
	git(t, root, "add", ".")
	git(t, root, "-c", "user.name=Test", "-c", "user.email=test@local", "commit", "-qm", "seed")
	ctx := context.Background()
	idx, _, err := Synced(ctx, root, false)
	if err != nil {
		t.Fatal(err)
	}
	defer idx.Close()
	links, err := idx.Links(ctx, "source")
	if err != nil || len(links.Links) != 2 {
		t.Fatalf("links: %+v %v", links, err)
	}
	for _, link := range links.Links {
		if link.TargetID == nil || *link.TargetID != "Café" {
			t.Fatalf("unresolved: %+v", link)
		}
	}
	result, _, err := Move(ctx, idx, root, "Cafe\u0301", "moved", false)
	if err != nil || !result.Commit || result.Total != 2 {
		t.Fatalf("move: %+v %v", result, err)
	}
	if status := git(t, root, "status", "--porcelain"); status != "" {
		t.Fatalf("unstaged Unicode path: %s", status)
	}
}
