package app

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCommitScopeAndIdentity(t *testing.T) {
	root := writeFixture(t)
	ctx := context.Background()
	idx, _, err := Synced(ctx, root, false)
	if err != nil {
		t.Fatal(err)
	}
	defer idx.Close()
	if _, err := Write(ctx, idx, root, "first", "one", "create", "new"); err != nil {
		t.Fatal(err)
	}
	if author := git(t, root, "log", "-1", "--format=%an <%ae>"); author != "mdkb <mdkb@local>" {
		t.Fatal(author)
	}
	git(t, root, "config", "user.name", "Test Identity")
	git(t, root, "config", "user.email", "test@example.invalid")
	if err := os.WriteFile(filepath.Join(root, "unrelated.md"), []byte("staged separately"), 0644); err != nil {
		t.Fatal(err)
	}
	git(t, root, "add", "--", "unrelated.md")
	if _, err := Write(ctx, idx, root, "second", "two", "create", "new"); err != nil {
		t.Fatal(err)
	}
	if author := git(t, root, "log", "-1", "--format=%an <%ae>"); author != "Test Identity <test@example.invalid>" {
		t.Fatal(author)
	}
	if files := git(t, root, "show", "--format=", "--name-only", "HEAD"); files != "second.md" {
		t.Fatal(files)
	}
	if staged := git(t, root, "diff", "--cached", "--name-only"); staged != "unrelated.md" {
		t.Fatal(staged)
	}
}

func TestMoveUntrackedSource(t *testing.T) {
	root := writeFixture(t)
	ctx := context.Background()
	idx, _, err := Synced(ctx, root, false)
	if err != nil {
		t.Fatal(err)
	}
	defer idx.Close()
	t.Setenv("MDKB_COMMIT_ON_WRITE", "0")
	if _, err := Write(ctx, idx, root, "old", "original", "create", "new"); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MDKB_COMMIT_ON_WRITE", "1")
	result, _, err := Move(ctx, idx, root, "old", "new", false)
	if err != nil || !result.Commit || result.CommitError != "" {
		t.Fatal(result, err)
	}
	if files := git(t, root, "show", "--format=", "--name-only", "HEAD"); !strings.Contains(files, "new.md") {
		t.Fatal(files)
	}
}
