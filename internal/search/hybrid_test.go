package search_test

import (
	"bytes"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/ContractorKeith/mdkb/internal/app"
	"github.com/ContractorKeith/mdkb/internal/embed"
	"github.com/ContractorKeith/mdkb/internal/embed/embedtest"
	"github.com/ContractorKeith/mdkb/internal/index"
	"github.com/ContractorKeith/mdkb/internal/search"
)

func TestHybridEndToEnd(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	root := t.TempDir()
	for id, body := range map[string]string{"cafe": "# Café\ncoffee coffee", "other": "# Other\nmountain hiking"} {
		if err := os.WriteFile(filepath.Join(root, id+".md"), []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}
	idx, err := index.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	defer idx.Close()
	if err = idx.Sync(t.Context(), root, false); err != nil {
		t.Fatal(err)
	}
	server := embedtest.Server(t)
	t.Setenv("MDKB_SEMANTIC", "1")
	t.Setenv("MDKB_OLLAMA_URL", server.URL)
	t.Setenv("MDKB_EMBED_MODEL", "test")
	e := embed.Ollama{URL: server.URL, ModelName: "test", Client: server.Client()}
	result, err := app.Embed(t.Context(), idx, e, false, false, nil)
	if err != nil || result.NotesEmbedded != 2 || result.Chunks != 2 {
		t.Fatalf("%+v %v", result, err)
	}
	result, err = app.Embed(t.Context(), idx, e, false, false, nil)
	if err != nil || result.Skipped != 2 || result.NotesEmbedded != 0 {
		t.Fatalf("%+v %v", result, err)
	}
	s := search.Search{Index: idx, Embedder: e}
	keyword, err := s.Keyword(t.Context(), "espresso", 3)
	if err != nil || len(keyword) != 0 {
		t.Fatal(keyword, err)
	}
	semantic, err := s.Semantic(t.Context(), "espresso", 3)
	if err != nil || semantic[0].ID != "cafe" {
		t.Fatal(semantic, err)
	}
	hybrid, err := s.Hybrid(t.Context(), "espresso", 3)
	if err != nil || hybrid[0].ID != "cafe" || !reflect.DeepEqual(hybrid[0].Sources, []string{"semantic"}) {
		t.Fatal(hybrid, err)
	}
	t.Logf("espresso: keyword=%v; semantic top=%s score=%.4f; hybrid top=%s score=%.4f sources=%v", keyword, semantic[0].ID, semantic[0].Score, hybrid[0].ID, hybrid[0].Score, hybrid[0].Sources)
	server.Close()
	var diagnostic bytes.Buffer
	fallback, err := app.SearchDefault(t.Context(), idx, "coffee", "", 3, &diagnostic)
	if err != nil || fallback.Mode != "keyword" || len(fallback.Hits) != 1 || !strings.Contains(diagnostic.String(), "keyword only") {
		t.Fatal(fallback, err, diagnostic.String())
	}
	t.Log(strings.TrimSpace(diagnostic.String()))
	if _, err := app.SearchDefault(t.Context(), idx, "coffee", "semantic", 3, nil); err == nil {
		t.Fatal("forced semantic succeeded")
	}
	for _, syncOne := range []bool{true, false} {
		if err = os.WriteFile(filepath.Join(root, "cafe.md"), []byte("# Changed\nnew larger content "+strings.Repeat("x", int(result.Skipped))), 0644); err != nil {
			t.Fatal(err)
		}
		if syncOne {
			err = idx.SyncNote(t.Context(), root, "cafe")
		} else {
			err = idx.Sync(t.Context(), root, false)
		}
		if err != nil {
			t.Fatal(err)
		}
		counts, err := idx.EmbeddingCounts(t.Context(), "test")
		if err != nil || counts.Notes != 1 {
			t.Fatal(counts, err)
		}
	}
	if err = idx.Sync(t.Context(), root, true); err != nil {
		t.Fatal(err)
	}
	counts, err := idx.EmbeddingCounts(t.Context(), "test")
	if err != nil || counts.Chunks != 0 {
		t.Fatal(counts, err)
	}
}
func TestRRFFusion(t *testing.T) {
	k := []search.Hit{{ID: "a"}, {ID: "b"}}
	s := []search.Hit{{ID: "b"}, {ID: "c"}}
	got := search.Fuse(k, s, 3)
	want := math.Round((1.0/62+1.0/61)*1e4) / 1e4
	if got[0].ID != "b" || got[0].Score != want || !reflect.DeepEqual(got[0].Sources, []string{"keyword", "semantic"}) {
		t.Fatal(got)
	}
	tied := search.Fuse([]search.Hit{{ID: "z"}}, []search.Hit{{ID: "a"}}, 2)
	if tied[0].ID != "a" {
		t.Fatal(tied)
	}
	t.Logf("b = round(1/62 + 1/61, 4) = %.4f; a = %.4f; c = %.4f", got[0].Score, got[1].Score, got[2].Score)
}
