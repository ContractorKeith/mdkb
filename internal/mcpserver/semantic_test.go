package mcpserver_test

import (
	"testing"

	"github.com/ContractorKeith/mdkb/internal/app"
	"github.com/ContractorKeith/mdkb/internal/embed/embedtest"
)

func TestSemanticSearchMode(t *testing.T) {
	root := fixture(t)
	server := embedtest.Server(t)
	t.Setenv("MDKB_SEMANTIC", "1")
	t.Setenv("MDKB_OLLAMA_URL", server.URL)
	t.Setenv("MDKB_EMBED_MODEL", "test")
	idx, _, err := app.Synced(t.Context(), root, false)
	if err != nil {
		t.Fatal(err)
	}
	s, _, err := app.Searcher(idx, "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = app.Embed(t.Context(), idx, s.Embedder, false, false, nil); err != nil {
		t.Fatal(err)
	}
	idx.Close()
	cs := client(t, root, false)
	for _, mode := range []string{"keyword", "semantic", "hybrid", ""} {
		got := decoded(t, success(t, cs, "search_notes", map[string]any{"query": "espresso", "mode": mode})).(map[string]any)
		expected := mode
		if mode == "" {
			expected = "hybrid"
		}
		if got["mode"] != expected {
			t.Fatal(got)
		}
		if mode != "keyword" {
			hits := got["hits"].([]any)
			if hits[0].(map[string]any)["id"] != "spaces and unicode" {
				t.Fatal(got)
			}
		}
	}
	if !call(t, cs, "search_notes", map[string]any{"query": "x", "mode": "bogus"}).IsError {
		t.Fatal("invalid mode")
	}
	server.Close()
	if !call(t, cs, "search_notes", map[string]any{"query": "x", "mode": "semantic"}).IsError {
		t.Fatal("semantic succeeded offline")
	}
	got := decoded(t, success(t, cs, "search_notes", map[string]any{"query": "shared"})).(map[string]any)
	if got["mode"] != "keyword" {
		t.Fatal(got)
	}
}
