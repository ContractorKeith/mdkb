package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	t.Setenv("MDKB_ROOT", root)
	if _, err := run("init", root); err != nil {
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
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(dest, data, 0644)
	})
	if err != nil {
		t.Fatal(err)
	}
	return root
}
func mustRun(t *testing.T, args ...string) string {
	t.Helper()
	out, err := run(args...)
	if err != nil {
		t.Fatalf("%v: %v", args, err)
	}
	return out
}
func TestGoldens(t *testing.T) {
	fixture(t)
	for _, tt := range []struct {
		name string
		args []string
	}{
		{"links-ambiguous-source", []string{"links", "ambiguous-source"}},
		{"links-code-fences", []string{"links", "code-fences"}},
		{"links-alias-source", []string{"links", "alias-source"}},
		{"links-heading-source", []string{"links", "heading-source"}},
		{"links-embed-source", []string{"links", "embed-source"}},
		{"links-path-source", []string{"links", "path-source"}},
		{"backlinks-plain", []string{"backlinks", "plain"}},
		{"orphans", []string{"orphans"}}, {"broken", []string{"broken"}},
		{"ls", []string{"ls"}}, {"tags", []string{"tags"}}, {"search-markdown", []string{"search", "markdown"}}, {"search-shared", []string{"search", "shared"}}, {"cat-welcome", []string{"cat", "welcome"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := mustRun(t, append(tt.args, "--json")...)
			path := filepath.Join("../../testdata/golden", tt.name+".json")
			if os.Getenv("UPDATE_GOLDEN") == "1" {
				if err := os.WriteFile(path, []byte(got), 0644); err != nil {
					t.Fatal(err)
				}
			}
			want, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if got != string(want) {
				t.Fatalf("golden differs:\n%s", got)
			}
		})
	}
}
func TestListCapAndRebuild(t *testing.T) {
	root := t.TempDir()
	t.Setenv("MDKB_ROOT", root)
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	mustRun(t, "init", root)
	for i := 0; i < 60; i++ {
		if err := os.WriteFile(filepath.Join(root, fmt.Sprintf("extra-%02d.md", i)), []byte("# Extra\nmarkdown\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	before := mustRun(t, "ls", "--json")
	var listing struct {
		Notes     []any
		Total     int
		Truncated bool
	}
	if err := json.Unmarshal([]byte(before), &listing); err != nil {
		t.Fatal(err)
	}
	if len(listing.Notes) != 50 || listing.Total != 60 || !listing.Truncated {
		t.Fatalf("bad cap: %+v", listing)
	}
	search := mustRun(t, "search", "markdown", "--json")
	var info struct {
		IndexPath string `json:"index_path"`
	}
	if err := json.Unmarshal([]byte(mustRun(t, "info", "--json")), &info); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(info.IndexPath); err != nil {
		t.Fatal(err)
	}
	mustRun(t, "reindex", "--full")
	if got := mustRun(t, "ls", "--json"); got != before {
		t.Fatal("rebuild changed listing")
	}
	if got := mustRun(t, "search", "markdown", "--json"); got != search {
		t.Fatal("rebuild changed search")
	}
}
func TestIncrementalCommands(t *testing.T) {
	root := fixture(t)
	mustRun(t, "ls")
	path := filepath.Join(root, "welcome.md")
	if err := os.WriteFile(path, []byte("---\ntags: unique\n---\n# Changed\nuniqueterm\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if got := mustRun(t, "search", "uniqueterm", "--json"); !strings.Contains(got, "Changed") {
		t.Fatal(got)
	}
	if got := mustRun(t, "ls", "--tag", "unique", "--json"); !strings.Contains(got, `"total":1`) {
		t.Fatal(got)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if got := mustRun(t, "search", "uniqueterm", "--json"); got != "{\"mode\":\"keyword\",\"hits\":[]}\n" {
		t.Fatal(got)
	}
	if got := mustRun(t, "tags", "--json"); strings.Contains(got, "unique") {
		t.Fatal(got)
	}
	for _, q := range []string{`unbalanced "quote`, "OR", "title:markdown", "(", "*", ""} {
		mustRun(t, "search", q, "--json")
	}
	for _, args := range [][]string{{"cat", "../outside"}, {"ls", "--limit", "0"}, {"search", "x", "--limit", "-1"}} {
		if _, err := run(args...); err == nil {
			t.Fatalf("expected error: %v", args)
		}
	}
}
func TestReadErrors(t *testing.T) {
	fixture(t)
	for _, root := range []string{t.TempDir(), filepath.Join(t.TempDir(), "missing")} {
		for _, command := range []string{"ls", "cat", "tags", "search", "reindex"} {
			args := []string{"--root", root, command}
			if command == "cat" || command == "search" {
				args = append(args, "x")
			}
			if out, err := run(args...); exitCode(err) != 3 || out != "" {
				t.Fatalf("%v: %q %v", args, out, err)
			}
		}
	}
	t.Setenv("XDG_CACHE_HOME", filepath.Join(t.TempDir(), "file"))
	if err := os.WriteFile(os.Getenv("XDG_CACHE_HOME"), []byte("block"), 0644); err != nil {
		t.Fatal(err)
	}
	if out, err := run("ls"); exitCode(err) != 4 || out != "" {
		t.Fatalf("index failure: %q %v", out, err)
	}
}
