package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestLinkReportsCapAndErrors(t *testing.T) {
	root := fixture(t)
	for _, command := range []string{"cat", "links", "backlinks"} {
		if out, err := run(command, "nope", "--json"); exitCode(err) != 2 || out != "" || err.Error() != "note not found: nope" {
			t.Fatalf("%s: %q %v", command, out, err)
		}
	}
	for n := 0; n < 60; n++ {
		path := filepath.Join(root, fmt.Sprintf("extra-%02d.md", n))
		if err := os.WriteFile(path, []byte("[[missing]]\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	for _, command := range []string{"orphans", "broken"} {
		before := mustRun(t, command, "--json")
		var data struct {
			Orphans   []any
			Broken    []any
			Total     int
			Truncated bool
		}
		if err := json.Unmarshal([]byte(before), &data); err != nil {
			t.Fatal(err)
		}
		if len(data.Orphans)+len(data.Broken) != 50 || data.Total <= 60 || !data.Truncated {
			t.Fatalf("%+v", data)
		}
		mustRun(t, "reindex", "--full")
		if got := mustRun(t, command, "--json"); got != before {
			t.Fatalf("rebuild changed %s", command)
		}
	}
}

func TestEmptyLinkGraph(t *testing.T) {
	root := t.TempDir()
	t.Setenv("MDKB_ROOT", root)
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	mustRun(t, "init", root)
	for _, command := range []string{"orphans", "broken"} {
		want := fmt.Sprintf("{\"%s\":[],\"total\":0,\"truncated\":false}\n", command)
		if got := mustRun(t, command, "--json"); got != want {
			t.Fatal(got)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "solo.md"), []byte("# Solo\n"), 0644); err != nil {
		t.Fatal(err)
	}
	for _, command := range []string{"links", "backlinks"} {
		want := fmt.Sprintf("{\"id\":\"solo\",\"%s\":[]}\n", command)
		if got := mustRun(t, command, "solo", "--json"); got != want {
			t.Fatal(got)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "solo.md"), []byte("[[solo]] ![[solo|self]]\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if got := mustRun(t, "orphans", "--json"); got != "{\"orphans\":[],\"total\":0,\"truncated\":false}\n" {
		t.Fatal(got)
	}
	var info struct {
		Links       int
		BrokenLinks int `json:"broken_links"`
		Orphans     int
	}
	if err := json.Unmarshal([]byte(mustRun(t, "info", "--json")), &info); err != nil {
		t.Fatal(err)
	}
	if info.Links != 2 || info.BrokenLinks != 0 || info.Orphans != 0 {
		t.Fatalf("%+v", info)
	}
}
