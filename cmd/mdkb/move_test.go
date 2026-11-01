package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ContractorKeith/mdkb/internal/kb"
)

func TestMoveFixture(t *testing.T) {
	root := writeFixture(t)
	gitKB(t, root, "add", "-A")
	gitKB(t, root, "-c", "user.name=t", "-c", "user.email=t@t", "commit", "-qm", "seed")
	before := mustRun(t, "broken", "--json")
	out, diff, err := cli(t, "", "mv", "plain", "renamed/plain", "--dry-run", "--json")
	if err != nil {
		t.Fatal(err)
	}
	for name, got := range map[string]string{"mv-dry-run.json": out, "mv-dry-run.diff": diff} {
		path := filepath.Join("../../testdata/golden", name)
		if os.Getenv("UPDATE_GOLDEN") == "1" {
			if err := os.WriteFile(path, []byte(got), 0644); err != nil {
				t.Fatal(err)
			}
		}
		want, err := os.ReadFile(path)
		if err != nil || got != string(want) {
			t.Fatalf("%s: %v\n%s", name, err, got)
		}
	}
	if status := gitKB(t, root, "status", "--porcelain"); status != "" {
		t.Fatal(status)
	}
	if _, err := os.Stat(filepath.Join(root, "renamed")); !os.IsNotExist(err) {
		t.Fatal("dry-run created directory")
	}
	got, stderr, err := cli(t, "", "mv", "plain", "renamed/plain", "--json")
	var result kb.MoveResult
	if err != nil || stderr != "" || json.Unmarshal([]byte(got), &result) != nil || !result.Commit || len(result.Rewritten) != 4 {
		t.Fatal(got, stderr, err)
	}
	if broken := mustRun(t, "broken", "--json"); broken != before {
		t.Fatal(broken, before)
	}
	backlinks := mustRun(t, "backlinks", "renamed/plain", "--json")
	for _, source := range []string{"welcome", "code-fences", "embed-source", "empty-frontmatter"} {
		if !strings.Contains(backlinks, source) {
			t.Fatal(backlinks)
		}
	}
	_, raw, err := kb.Read(root, "code-fences")
	if err != nil || !strings.Contains(raw, "[[fenced-missing]]") || !strings.Contains(raw, "[[renamed/plain]]") {
		t.Fatal(raw, err)
	}
	if count := gitKB(t, root, "rev-list", "--count", "HEAD"); count != "2" {
		t.Fatal(count)
	}
	if msg := gitKB(t, root, "log", "-1", "--format=%s"); msg != "mdkb: mv plain -> renamed/plain (4 links rewritten)" {
		t.Fatal(msg)
	}
	if _, _, err := kb.Read(root, "plain"); !errors.Is(err, os.ErrNotExist) {
		t.Fatal(err)
	}
}

func TestDailyConfig(t *testing.T) {
	root := writeFixture(t)
	home := os.Getenv("HOME")
	if err := os.MkdirAll(filepath.Join(home, ".config/mdkb"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".config/mdkb/config.toml"), []byte("[write]\ndaily_dir = 'Journal'\n"), 0644); err != nil {
		t.Fatal(err)
	}
	mustRun(t, "daily", "--date", "2026-02-03", "--json")
	if _, raw, err := kb.Read(root, "Journal/2026-02-03"); err != nil || !strings.Contains(raw, "# 2026-02-03") {
		t.Fatal(raw, err)
	}
}
