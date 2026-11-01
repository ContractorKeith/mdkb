package app

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ContractorKeith/mdkb/internal/kb"
)

func git(t *testing.T, root string, args ...string) string {
	t.Helper()
	out, err := exec.Command("git", append([]string{"-C", root}, args...)...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %s %v", args, out, err)
	}
	return strings.TrimSpace(string(out))
}
func writeFixture(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(home, "gitconfig"))
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	t.Setenv("MDKB_COMMIT_ON_WRITE", "1")
	root := filepath.Join(t.TempDir(), "kb")
	if err := kb.Init(root); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestWriteCommits(t *testing.T) {
	root := writeFixture(t)
	ctx := context.Background()
	idx, _, err := Synced(ctx, root, false)
	if err != nil {
		t.Fatal(err)
	}
	defer idx.Close()
	for _, step := range []struct{ mode, body string }{{"create", "first"}, {"append", "second"}, {"overwrite", "third"}} {
		got, err := Write(ctx, idx, root, "welcome", step.body, step.mode, step.mode)
		if err != nil || !got.Commit || got.CommitError != "" {
			t.Fatalf("%+v %v", got, err)
		}
		if message := git(t, root, "log", "-1", "--format=%s"); message != "mdkb: "+map[string]string{"create": "new", "append": "append", "overwrite": "overwrite"}[step.mode]+" welcome" {
			t.Fatal(message)
		}
	}
	if count := git(t, root, "rev-list", "--count", "HEAD"); count != "3" {
		t.Fatal(count)
	}
	got, err := Write(ctx, idx, root, "welcome", "third", "overwrite", "overwrite")
	if err != nil || got.Commit {
		t.Fatal(got, err)
	}
	if count := git(t, root, "rev-list", "--count", "HEAD"); count != "3" {
		t.Fatal(count)
	}
	t.Setenv("MDKB_COMMIT_ON_WRITE", "0")
	got, err = Write(ctx, idx, root, "uncommitted", "x", "create", "new")
	if err != nil || got.Commit {
		t.Fatal(got, err)
	}
	if status := git(t, root, "status", "--short"); !strings.Contains(status, "?? uncommitted.md") {
		t.Fatal(status)
	}
	// A hook failure reports commit:false and preserves the published note.
	t.Setenv("MDKB_COMMIT_ON_WRITE", "1")
	if err := os.WriteFile(filepath.Join(root, ".git/hooks/pre-commit"), []byte("#!/bin/sh\nexit 1\n"), 0755); err != nil {
		t.Fatal(err)
	}
	got, err = Write(ctx, idx, root, "hook-failure", "x", "create", "new")
	if err != nil || got.Commit || got.CommitError == "" {
		t.Fatal(got, err)
	}
	if _, _, err := kb.Read(root, "hook-failure"); err != nil {
		t.Fatal(err)
	}
}

func TestNonRepoAndConfig(t *testing.T) {
	root := writeFixture(t)
	if err := os.RemoveAll(filepath.Join(root, ".git")); err != nil {
		t.Fatal(err)
	}
	if _, err := commitWrite(root, []string{"missing.md"}, "mdkb: new missing"); err != nil {
		t.Fatal(err)
	}
	home := os.Getenv("HOME")
	if err := os.MkdirAll(filepath.Join(home, ".config/mdkb"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".config/mdkb/config.toml"), []byte("[write]\ncommit_on_write = false\ndaily_dir = 'Journal'\n"), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MDKB_COMMIT_ON_WRITE", "")
	git(t, root, "init", "-q")

	if _, err := commitWrite(root, []string{"missing.md"}, "mdkb: new missing"); err != nil {
		t.Fatal(err)
	}
}

func TestRestoreCommittedContent(t *testing.T) {
	root := writeFixture(t)
	ctx := context.Background()
	idx, _, err := Synced(ctx, root, false)
	if err != nil {
		t.Fatal(err)
	}
	defer idx.Close()
	if _, err := Write(ctx, idx, root, "note", "original", "create", "new"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "note.md"), []byte("manual change\n"), 0644); err != nil {
		t.Fatal(err)
	}
	result, err := Write(ctx, idx, root, "note", "original", "overwrite", "overwrite")
	if err != nil || result.Commit || result.CommitError != "" {
		t.Fatal(result, err)
	}
	if count := git(t, root, "rev-list", "--count", "HEAD"); count != "1" {
		t.Fatal(count)
	}
}
