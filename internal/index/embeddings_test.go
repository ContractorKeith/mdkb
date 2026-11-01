package index

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestEmbeddingStorageLifecycle(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	root := t.TempDir()
	path := filepath.Join(root, "a.md")
	if err := os.WriteFile(path, []byte("# A\nbody"), 0644); err != nil {
		t.Fatal(err)
	}
	idx, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	defer idx.Close()
	if err = idx.Sync(t.Context(), root, false); err != nil {
		t.Fatal(err)
	}
	notes, err := idx.EmbedNotes(t.Context(), "m")
	if err != nil || len(notes) != 1 || notes[0].Current {
		t.Fatal(notes, err)
	}
	vectors := [][]float32{{1.25, -2, 3}, {0, 1, 0}}
	if err = idx.StoreEmbeddings(t.Context(), notes[0], "m", vectors); err != nil {
		t.Fatal(err)
	}
	stored, err := idx.Vectors(t.Context(), "m")
	if err != nil || len(stored) != 2 || !reflect.DeepEqual(stored[0].Values, vectors[0]) {
		t.Fatal(stored, err)
	}
	// Failure after the first insert must roll back the delete and all inserts.
	if err = idx.StoreEmbeddings(t.Context(), notes[0], "m", [][]float32{{9, 9, 9}, {1}}); err == nil {
		t.Fatal("accepted inconsistent dimensions")
	}
	stored, err = idx.Vectors(t.Context(), "m")
	if err != nil || !reflect.DeepEqual(stored[0].Values, vectors[0]) || len(stored) != 2 {
		t.Fatal(stored, err)
	}
	current, err := idx.EmbedNotes(t.Context(), "m")
	if err != nil || !current[0].Current {
		t.Fatal(current, err)
	}
	other, err := idx.EmbedNotes(t.Context(), "other")
	if err != nil || other[0].Current {
		t.Fatal(other, err)
	}
	if err = idx.SyncNote(t.Context(), root, "a"); err != nil {
		t.Fatal(err)
	}
	counts, err := idx.EmbeddingCounts(t.Context(), "m")
	if err != nil || counts.Chunks != 2 {
		t.Fatal(counts, err)
	}
	if err = os.WriteFile(path, []byte("# Changed\nnew body"), 0644); err != nil {
		t.Fatal(err)
	}
	if err = idx.Sync(t.Context(), root, false); err != nil {
		t.Fatal(err)
	}
	counts, err = idx.EmbeddingCounts(t.Context(), "m")
	if err != nil || counts.Chunks != 0 {
		t.Fatal(counts, err)
	}
	if err = idx.StoreEmbeddings(t.Context(), notes[0], "m", vectors); err == nil {
		t.Fatal("accepted stale stamp")
	}
	notes, err = idx.EmbedNotes(t.Context(), "m")
	if err != nil {
		t.Fatal(err)
	}
	if err = idx.StoreEmbeddings(t.Context(), notes[0], "m", vectors); err != nil {
		t.Fatal(err)
	}
	if err = os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err = idx.Sync(t.Context(), root, false); err != nil {
		t.Fatal(err)
	}
	counts, err = idx.EmbeddingCounts(t.Context(), "m")
	if err != nil || counts.Chunks != 0 {
		t.Fatal(counts, err)
	}
	notes, err = idx.EmbedNotes(t.Context(), "m")
	if err != nil || len(notes) != 0 {
		t.Fatal(notes, err)
	}
}
