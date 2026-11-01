package kb

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMoveSelfLinkAndNoFinalNewline(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(home, "gitconfig"))
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	root := filepath.Join(t.TempDir(), "kb")
	if err := Init(root); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "old.md"), []byte("[[old#h|a]]"), 0644); err != nil {
		t.Fatal(err)
	}
	plan, err := PlanMove(root, "old", "nested/new", nil)
	if err != nil || plan.Result.Total != 1 {
		t.Fatal(plan, err)
	}
	if err := ApplyMove(root, plan); err != nil {
		t.Fatal(err)
	}
	_, raw, err := Read(root, "nested/new")
	if err != nil || raw != "[[nested/new#h|a]]\n" {
		t.Fatal(raw, err)
	}
	if _, _, err := Read(root, "old"); !errors.Is(err, os.ErrNotExist) {
		t.Fatal(err)
	}
}

func TestMoveRefusesSymlink(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(home, "gitconfig"))
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	root := filepath.Join(t.TempDir(), "kb")
	if err := Init(root); err != nil {
		t.Fatal(err)
	}
	if _, err := Write(root, "old", "original", "create"); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(root, "old.md"), filepath.Join(root, "linked.md")); err != nil {
		t.Fatal(err)
	}
	for _, pair := range [][2]string{{"linked", "new"}, {"old", "linked"}, {"old", "old"}, {"old", "../new"}} {
		_, err := PlanMove(root, pair[0], pair[1], nil)
		if !errors.Is(err, ErrWriteRefused) {
			t.Fatal(pair, err)
		}
	}
	_, raw, err := Read(root, "old")
	if err != nil || !strings.Contains(raw, "original") {
		t.Fatal(raw, err)
	}
}
