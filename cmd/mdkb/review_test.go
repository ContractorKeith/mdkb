package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMalformedNotesRemainUsable(t *testing.T) {
	root := fixture(t)
	// Include invalid YAML as well as the fixture's unterminated block.
	raw := "---\ntitle: Broken metadata\ntags: [unfinished\n---\n# Invalid YAML fallback\nrecoverabletoken\n"
	path := filepath.Join(root, "invalid.md")
	if err := os.WriteFile(path, []byte(raw), 0644); err != nil {
		t.Fatal(err)
	}
	listing := mustRun(t, "ls", "--json")
	var data struct {
		Notes []struct {
			ID, Problem string
			Tags        []string
		}
		Total int
	}
	if err := json.Unmarshal([]byte(listing), &data); err != nil {
		t.Fatal(err)
	}
	problems := 0
	for _, n := range data.Notes {
		if n.Problem != "" {
			problems++
			if len(n.Tags) != 0 {
				t.Fatalf("fallback tags: %+v", n)
			}
		}
	}
	if data.Total != 23 || problems != 2 {
		t.Fatalf("total=%d problems=%d", data.Total, problems)
	}
	var info struct{ Problems int }
	if err := json.Unmarshal([]byte(mustRun(t, "info", "--json")), &info); err != nil {
		t.Fatal(err)
	}
	if info.Problems != 2 {
		t.Fatalf("info: %+v", info)
	}
	if got := mustRun(t, "search", "recoverabletoken", "--json"); !strings.Contains(got, "invalid") {
		t.Fatal(got)
	}
	if got := mustRun(t, "cat", "invalid"); got != raw {
		t.Fatal(got)
	}
	mustRun(t, "cat", "welcome")
	mustRun(t, "tags")
	// A repaired file clears the stored problem and updates the inspector count.
	if err := os.WriteFile(path, []byte("# Repaired\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if got := mustRun(t, "ls", "--json"); strings.Contains(got, "invalid frontmatter") {
		t.Fatal(got)
	}
	if err := json.Unmarshal([]byte(mustRun(t, "info", "--json")), &info); err != nil {
		t.Fatal(err)
	}
	if info.Problems != 1 {
		t.Fatalf("repaired: %+v", info)
	}
}

func TestCatRefusesSymlink(t *testing.T) {
	root := fixture(t)
	outside := filepath.Join(t.TempDir(), "outside.md")
	if err := os.WriteFile(outside, []byte("private outside contents"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "outside-link.md")); err != nil {
		t.Fatal(err)
	}
	out, err := run("cat", "outside-link")
	if exitCode(err) != 2 || out != "" {
		t.Fatalf("symlink cat: %q, %v", out, err)
	}
}
