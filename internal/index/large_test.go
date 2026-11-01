package index

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLargeUnchangedList(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	root := t.TempDir()
	for n := 0; n < 5000; n++ {
		body := fmt.Sprintf("---\naliases: [alias-%d]\n---\n# Note %d\n[[note-%04d]] [[alias-%d]]\n", n, n, (n+1)%5000, (n+2)%5000)
		if err := os.WriteFile(filepath.Join(root, fmt.Sprintf("note-%04d.md", n)), []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}
	idx, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	defer idx.Close()
	ctx := context.Background()
	if err := idx.Sync(ctx, root, false); err != nil {
		t.Fatal(err)
	}
	// A trigger proves unchanged sync never attempts graph updates, even no-op updates.
	if _, err := idx.db.Exec(`CREATE TRIGGER forbid_resolution BEFORE UPDATE ON links BEGIN SELECT RAISE(ABORT,'unchanged graph updated'); END`); err != nil {
		t.Fatal(err)
	}
	start := time.Now()
	if err := idx.Sync(ctx, root, false); err != nil {
		t.Fatal(err)
	}
	listing, err := idx.List(ctx, "", 50)
	elapsed := time.Since(start)
	if err != nil || listing.Total != 5000 || len(listing.Notes) != 50 {
		t.Fatalf("listing: %+v %v", listing, err)
	}
	t.Logf("unchanged sync + ls on 5,000 linked notes: %s", elapsed)
	if elapsed >= 2*time.Second {
		t.Fatalf("unchanged ls took %s; want <2s", elapsed)
	}
}
