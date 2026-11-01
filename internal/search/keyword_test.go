package search

import (
	"context"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/ContractorKeith/mdkb/internal/index"
)

func TestKeywordRoundsPublicScores(t *testing.T) {
	root := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	for name, body := range map[string]string{"a.md": "# Search\nneedle needle\n", "b.md": "# Other\nneedle and filler words\n", "c.md": "# Unrelated\ntext\n", "d.md": "# Another\nfiller\n", "e.md": "# Last\nmore filler\n"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}
	idx, err := index.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	defer idx.Close()
	ctx := context.Background()
	if err := idx.Sync(ctx, root, false); err != nil {
		t.Fatal(err)
	}
	raw, err := idx.Keyword(ctx, "needle", 50)
	if err != nil {
		t.Fatal(err)
	}
	hits, err := (Search{Index: idx}).Keyword(ctx, "needle", 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 2 {
		t.Fatalf("%+v", hits)
	}
	for i, h := range hits {
		if h.ID != raw[i].ID || h.Score != math.Round(raw[i].Score*1e4)/1e4 {
			t.Fatalf("score/order mismatch: %+v raw %+v", h, raw[i])
		}
	}
}
