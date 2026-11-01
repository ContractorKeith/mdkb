package mcpserver_test

import (
	"context"
	"fmt"
	"github.com/ContractorKeith/mdkb/internal/index"
	"github.com/ContractorKeith/mdkb/internal/kb"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWrites(t *testing.T) {
	root := fixture(t)
	cs := client(t, root, true)
	write := func(id, body, mode string) any {
		return success(t, cs, "write_note", map[string]any{"id": id, "body": body, "mode": mode})
	}
	for _, tt := range []struct{ mode, body, want string }{
		{"create", "firsttoken", "firsttoken\n"},
		{"append", "secondtoken", "firsttoken\n\nsecondtoken\n"},
		{"overwrite", "replacementtoken", "replacementtoken\n"},
	} {
		got := decoded(t, write("nested/new", tt.body, tt.mode)).(map[string]any)
		if got["commit"] != true || got["id"] != "nested/new" || got["path"] != "nested/new.md" || got["bytes"] != float64(len(tt.want)) || got["mode"] != tt.mode {
			t.Fatal(got)
		}
		raw, err := os.ReadFile(filepath.Join(root, "nested/new.md"))
		if err != nil || string(raw) != tt.want {
			t.Fatalf("%q %v", raw, err)
		}
		idx, err := index.Open(root)
		if err != nil {
			t.Fatal(err)
		}
		matches, err := idx.Keyword(context.Background(), tt.body, 50)
		idx.Close()
		if err != nil || len(matches) != 1 || matches[0].ID != "nested/new" {
			t.Fatalf("write did not reindex: %+v %v", matches, err)
		}
		hits := decoded(t, success(t, cs, "search_notes", map[string]any{"query": tt.body})).(map[string]any)["hits"].([]any)
		if len(hits) != 1 || hits[0].(map[string]any)["id"] != "nested/new" {
			t.Fatal(hits)
		}
	}
	if !call(t, cs, "write_note", map[string]any{"id": "nested/new", "body": "lost", "mode": "create"}).IsError {
		t.Fatal("create clobbered")
	}
	for _, args := range []map[string]any{
		{"id": "../x", "body": "bad", "mode": "create"},
		{"id": ".hidden/x", "body": "bad", "mode": "create"},
		{"id": "nested/./bad", "body": "bad", "mode": "create"},
		{"id": "missing", "body": "bad", "mode": "append"},
		{"id": "missing", "body": "bad", "mode": "overwrite"},
		{"id": "missing", "body": "bad", "mode": "invalid"},
		{"id": "nested/new", "body": "bad", "mode": "append", "frontmatter": map[string]any{"title": "Bad"}},
	} {
		if !call(t, cs, "write_note", args).IsError {
			t.Fatalf("accepted %v", args)
		}
	}
	for _, path := range []string{filepath.Join(filepath.Dir(root), "x.md"), filepath.Join(root, ".hidden"), filepath.Join(root, "nested/bad.md")} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("unexpected write %s: %v", path, err)
		}
	}
	success(t, cs, "write_note", map[string]any{"id": "meta", "body": "metadataunique", "mode": "create", "frontmatter": map[string]any{"title": "Title: YAML", "tags": []string{"newtag"}, "aliases": []string{"Alias"}}})
	n, _, err := kb.Read(root, "meta")
	if err != nil || n.Title != "Title: YAML" || len(n.Tags) != 1 || n.Tags[0] != "newtag" {
		t.Fatalf("%+v %v", n, err)
	}
	captured := decoded(t, success(t, cs, "capture", map[string]any{"text": "captureuniquetoken", "title": "A Title!"})).(map[string]any)
	id := captured["id"].(string)
	if !strings.HasPrefix(id, "00-Inbox/") || !strings.HasSuffix(id, "-a-title") {
		t.Fatal(id)
	}
	n, _, err = kb.Read(root, id)
	if err != nil || n.Created == "" || n.Title != "A Title!" || n.Body != "captureuniquetoken\n" {
		t.Fatalf("%+v %v", n, err)
	}
	hits := decoded(t, success(t, cs, "search_notes", map[string]any{"query": "captureuniquetoken"})).(map[string]any)["hits"].([]any)
	if len(hits) != 1 {
		t.Fatal(hits)
	}
	// The write's single-note reindex also updates links and aliases.
	write("linker", "[[Alias]]", "create")
	links := decoded(t, success(t, cs, "get_links", map[string]any{"id": "linker"}))
	if !strings.Contains(fmt.Sprint(links), "target_id:meta") {
		t.Fatal(links)
	}
}

func TestReadCapsAndProblems(t *testing.T) {
	root := fixture(t)
	cs := client(t, root, true)
	for i := 0; i < 60; i++ {
		success(t, cs, "write_note", map[string]any{"id": fmt.Sprintf("extra-%02d", i), "body": "capuniquetoken", "mode": "create"})
	}
	list := decoded(t, success(t, cs, "list_notes", map[string]any{"limit": 100})).(map[string]any)
	if len(list["notes"].([]any)) != 50 || list["truncated"] != true {
		t.Fatal(list)
	}
	hits := decoded(t, success(t, cs, "search_notes", map[string]any{"query": "capuniquetoken", "limit": 100})).(map[string]any)["hits"].([]any)
	if len(hits) != 50 {
		t.Fatal(len(hits))
	}
	for _, name := range []string{"list_notes", "search_notes"} {
		for _, limit := range []int{0, -1} {
			if !call(t, cs, name, map[string]any{"limit": limit, "query": "capuniquetoken"}).IsError {
				t.Fatal("accepted invalid limit")
			}
		}
	}
	success(t, cs, "write_note", map[string]any{"id": "malformed", "body": "---\ntags: [\n---\n# Recovered\n", "mode": "create"})
	note := decoded(t, success(t, cs, "read_note", map[string]any{"id": "malformed"})).(map[string]any)
	if note["problem"] == nil {
		t.Fatal(note)
	}
}
