package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ContractorKeith/mdkb/internal/kb"
)

func writeFixture(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(home, "gitconfig"))
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("MDKB_COMMIT_ON_WRITE", "1")
	return fixture(t)
}
func cli(t *testing.T, input string, args ...string) (string, string, error) {
	t.Helper()
	cmd := newRoot()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetIn(strings.NewReader(input))
	cmd.SetArgs(args)
	err := cmd.Execute()
	return stdout.String(), stderr.String(), err
}
func gitKB(t *testing.T, root string, args ...string) string {
	t.Helper()
	out, err := exec.Command("git", append([]string{"-C", root}, args...)...).CombinedOutput()
	if err != nil {
		t.Fatalf("%v: %s %v", args, out, err)
	}
	return strings.TrimSpace(string(out))
}

func TestCLIWriteCreationAndDryRun(t *testing.T) {
	root := writeFixture(t)
	for _, tt := range []struct {
		args      []string
		input, id string
	}{
		{[]string{"new", "Meeting Notes", "--tags", "work,meeting", "--body", "agenda"}, "", "meeting-notes"},
		{[]string{"new", "From Stdin", "--stdin", "--dir", "projects"}, "stdin text", "projects/from-stdin"},
		{[]string{"capture", "remember", "the", "milk", "--title", "Groceries"}, "", ""},
		{[]string{"daily", "--date", "2026-11-01", "--append", "did a thing"}, "", "10-Daily/2026-11-01"},
	} {
		before := gitKB(t, root, "status", "--porcelain")
		args := append(append([]string(nil), tt.args...), "--dry-run", "--json")
		out, preview, err := cli(t, tt.input, args...)
		var dry struct {
			ID, Path string
			DryRun   bool `json:"dry_run"`
		}
		if err != nil || json.Unmarshal([]byte(out), &dry) != nil || !dry.DryRun || preview == "" {
			t.Fatalf("%s %s %v", out, preview, err)
		}
		if after := gitKB(t, root, "status", "--porcelain"); after != before {
			t.Fatal("dry-run changed KB")
		}
		out, stderr, err := cli(t, tt.input, append(tt.args, "--json")...)
		var result kb.WriteResult
		if err != nil || stderr != "" || json.Unmarshal([]byte(out), &result) != nil || !result.Commit {
			t.Fatalf("%s %s %v", out, stderr, err)
		}
		if tt.id != "" && result.ID != tt.id {
			t.Fatal(result.ID)
		}
		note, _, err := kb.Read(root, result.ID)
		if err != nil || note.Created == "" {
			t.Fatal(note, err)
		}
		if tt.id == "meeting-notes" && (len(note.Tags) != 2 || note.Title != "Meeting Notes" || note.Body != "agenda\n") {
			t.Fatal(note)
		}
	}
	if count := gitKB(t, root, "rev-list", "--count", "HEAD"); count != "4" {
		t.Fatal(count)
	}
	before := gitKB(t, root, "rev-parse", "HEAD")
	mustRun(t, "daily", "--date", "2026-11-01", "--json")
	if gitKB(t, root, "rev-parse", "HEAD") != before {
		t.Fatal("daily read committed")
	}
	mustRun(t, "daily", "--date", "2026-11-01", "--append", "second", "--json")
	_, raw, err := kb.Read(root, "10-Daily/2026-11-01")
	if err != nil || !strings.Contains(raw, "did a thing\n\nsecond\n") {
		t.Fatal(raw, err)
	}
}

func TestWriteRefusalsAndCommitWarning(t *testing.T) {
	root := writeFixture(t)
	mustRun(t, "new", "Existing")
	for _, tt := range []struct {
		args []string
		code int
	}{
		{[]string{"new", "Existing"}, 5},
		{[]string{"new", "Bad", "--dir", "../escape"}, 5},
		{[]string{"mv", "plain", "welcome"}, 5},
		{[]string{"mv", "nope", "elsewhere"}, 2},
		{[]string{"mv", "../plain", "elsewhere"}, 5},
		{[]string{"mv", "plain", "../elsewhere"}, 5},
		{[]string{"new", "Bad", "--root", t.TempDir()}, 5},
		{[]string{"daily", "--date", "bad"}, 1},
		{[]string{"new", "Bad", "--body", "x", "--stdin"}, 1},
	} {
		if _, _, err := cli(t, "", tt.args...); exitCode(err) != tt.code {
			t.Fatalf("%v: %v (%d)", tt.args, err, exitCode(err))
		}
	}
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "linked")); err != nil {
		t.Fatal(err)
	}
	if _, _, err := cli(t, "", "mv", "plain", "linked/plain"); exitCode(err) != 5 {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".git/hooks/pre-commit"), []byte("#!/bin/sh\nexit 1\n"), 0755); err != nil {
		t.Fatal(err)
	}
	out, stderr, err := cli(t, "", "new", "Survives", "--body", "x", "--json")
	if err != nil || !strings.Contains(out, `"commit":false`) || !strings.Contains(stderr, "git commit") {
		t.Fatal(out, stderr, err)
	}
	if _, _, err := kb.Read(root, "survives"); err != nil {
		t.Fatal(err)
	}
}
