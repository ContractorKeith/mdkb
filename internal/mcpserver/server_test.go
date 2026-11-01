package mcpserver_test

import (
	"context"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/ContractorKeith/mdkb/internal/kb"
	"github.com/ContractorKeith/mdkb/internal/mcpserver"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func fixture(t *testing.T) string {
	t.Helper()
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(home, "gitconfig"))
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("MDKB_COMMIT_ON_WRITE", "1")
	root := t.TempDir()
	if err := kb.Init(root); err != nil {
		t.Fatal(err)
	}
	err := filepath.WalkDir("../../testdata/kb", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel("../../testdata/kb", path)
		if err != nil {
			return err
		}
		dest := filepath.Join(root, rel)
		if d.IsDir() {
			return os.MkdirAll(dest, 0755)
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(dest, raw, 0644)
	})
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func client(t *testing.T, root string, writes bool) *mcp.ClientSession {
	t.Helper()
	server, err := mcpserver.New(root, writes)
	if err != nil {
		t.Fatal(err)
	}
	a, b := mcp.NewInMemoryTransports()
	ss, err := server.Connect(t.Context(), a, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ss.Close() })
	cs, err := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "0"}, nil).Connect(t.Context(), b, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { cs.Close() })
	return cs
}

func call(t *testing.T, cs *mcp.ClientSession, name string, args any) *mcp.CallToolResult {
	t.Helper()
	result, err := cs.CallTool(t.Context(), &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func success(t *testing.T, cs *mcp.ClientSession, name string, args any) any {
	t.Helper()
	r := call(t, cs, name, args)
	if r.IsError {
		t.Fatalf("%s: %+v", name, r.Content)
	}
	return r.StructuredContent
}

func decoded(t *testing.T, v any) any {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	var result any
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatal(err)
	}
	return result
}

func golden(t *testing.T, name string, got any) {
	t.Helper()
	raw, err := os.ReadFile("../../testdata/golden/" + name + ".json")
	if err != nil {
		t.Fatal(err)
	}
	var want any
	if err := json.Unmarshal(raw, &want); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(decoded(t, got), want) {
		t.Fatalf("%s: got %+v; want %+v", name, got, want)
	}
}

func TestToolsAndGoldens(t *testing.T) {
	root := fixture(t)
	for _, writes := range []bool{false, true} {
		cs := client(t, root, writes)
		list, err := cs.ListTools(context.Background(), nil)
		if err != nil {
			t.Fatal(err)
		}
		var got []string
		for _, tool := range list.Tools {
			got = append(got, tool.Name)
		}
		want := []string{"get_backlinks", "get_info", "get_links", "list_notes", "read_note", "search_notes"}
		if writes {
			want = append(want, "capture", "write_note")
		}
		sort.Strings(got)
		sort.Strings(want)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("tools: %v != %v", got, want)
		}
		for _, tt := range []struct {
			name, golden string
			args         map[string]any
		}{
			{"list_notes", "ls", map[string]any{}},
			{"read_note", "cat-welcome", map[string]any{"id": "welcome"}},
			{"search_notes", "search-shared", map[string]any{"query": "shared"}},
			{"search_notes", "search-markdown", map[string]any{"query": "markdown"}},
			{"get_links", "links-path-source", map[string]any{"id": "path-source"}},
			{"get_links", "links-ambiguous-source", map[string]any{"id": "ambiguous-source"}},
			{"get_backlinks", "backlinks-plain", map[string]any{"id": "plain"}},
		} {
			golden(t, tt.golden, success(t, cs, tt.name, tt.args))
		}
		info := decoded(t, success(t, cs, "get_info", map[string]any{})).(map[string]any)
		if info["root"] != root || info["index_path"] == "" {
			t.Fatal(info)
		}
		info["root"], info["index_path"] = "<root>", "<index>"
		golden(t, "info", info)
		for _, name := range []string{"read_note", "get_links", "get_backlinks"} {
			result := call(t, cs, name, map[string]any{"id": "unknown-note"})
			if !result.IsError || len(result.Content) != 1 {
				t.Fatalf("%s: %+v", name, result)
			}
			text, ok := result.Content[0].(*mcp.TextContent)
			if !ok || text.Text != "note not found: unknown-note" {
				t.Fatalf("%s: %+v", name, result.Content)
			}
		}
		if !writes {
			r, err := cs.CallTool(t.Context(), &mcp.CallToolParams{Name: "write_note", Arguments: map[string]any{"id": "forbidden", "body": "x", "mode": "create"}})
			if err == nil && !r.IsError {
				t.Fatal("write accepted without flag")
			}
			if _, err := os.Stat(filepath.Join(root, "forbidden.md")); !os.IsNotExist(err) {
				t.Fatal(err)
			}
		}
	}
}
