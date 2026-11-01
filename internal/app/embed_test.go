package app_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ContractorKeith/mdkb/internal/app"
	"github.com/ContractorKeith/mdkb/internal/embed"
	"github.com/ContractorKeith/mdkb/internal/index"
)

func TestEmbedLaterBatchFailureIsAtomic(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "long.md"), []byte(strings.Repeat("word ", 3000)), 0644); err != nil {
		t.Fatal(err)
	}
	idx, err := index.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	defer idx.Close()
	if err = idx.Sync(t.Context(), root, false); err != nil {
		t.Fatal(err)
	}
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if requests == 2 {
			http.Error(w, `{"error":"model unloaded"}`, 404)
			return
		}
		var input struct{ Input []string }
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			t.Error(err)
		}
		vectors := [][]float32{}
		for range input.Input {
			vectors = append(vectors, []float32{1, 2})
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"embeddings": vectors})
	}))
	defer server.Close()
	e := embed.Ollama{URL: server.URL, ModelName: "test", Client: server.Client()}
	if _, err = app.Embed(t.Context(), idx, e, false, false, nil); err == nil {
		t.Fatal("expected second batch error")
	}
	counts, err := idx.EmbeddingCounts(t.Context(), "test")
	if err != nil || counts.Chunks != 0 || requests != 2 {
		t.Fatal(counts, requests, err)
	}
	notes, err := idx.EmbedNotes(t.Context(), "test")
	if err != nil || notes[0].Current {
		t.Fatal(notes, err)
	}
}
