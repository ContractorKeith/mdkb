package kb_test

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/ContractorKeith/mdkb/internal/kb"
)

func write(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestInit(t *testing.T) {
	root := filepath.Join(t.TempDir(), "new")
	if err := kb.Init(root); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"00-Inbox", ".mdkb/config.toml", ".git"} {
		if _, err := os.Stat(filepath.Join(root, path)); err != nil {
			t.Fatal(err)
		}
	}
	ignore, err := os.ReadFile(filepath.Join(root, ".gitignore"))
	if err != nil || string(ignore) != ".mdkb/\n.DS_Store\n" {
		t.Fatalf("gitignore = %q, %v", ignore, err)
	}
	marker := filepath.Join(root, ".mdkb/config.toml")
	before, err := os.Stat(marker)
	if err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(root, ".gitignore"), "custom\n")
	write(t, filepath.Join(root, "note.md"), "# Note")
	if err := kb.Init(root); err != nil {
		t.Fatal(err)
	}
	after, err := os.Stat(marker)
	if err != nil {
		t.Fatal(err)
	}
	if !before.ModTime().Equal(after.ModTime()) {
		t.Fatal("second init rewrote marker")
	}
	ignore, err = os.ReadFile(filepath.Join(root, ".gitignore"))
	if err != nil || string(ignore) != "custom\n" {
		t.Fatal("second init changed gitignore")
	}
	info, err := kb.Info(root)
	if err != nil || !info.Initialized || info.NoteCount != 1 {
		t.Fatalf("info = %+v, %v", info, err)
	}
}

func TestInitRefusesNonKB(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "x"), "keep")
	for _, init := range []func(string) error{kb.Init, kb.PreviewInit} {
		if err := init(root); !errors.Is(err, kb.ErrWriteRefused) {
			t.Fatalf("got %v", err)
		}
	}
	entries, err := os.ReadDir(root)
	if err != nil || len(entries) != 1 {
		t.Fatalf("refused init changed directory: %v", err)
	}
	file := filepath.Join(root, "x")
	if err := kb.Init(file); !errors.Is(err, kb.ErrWriteRefused) {
		t.Fatalf("file: %v", err)
	}
}

func TestInitInsideRepo(t *testing.T) {
	parent := t.TempDir()
	if output, err := exec.Command("git", "-C", parent, "init", "--quiet").CombinedOutput(); err != nil {
		t.Fatalf("%v: %s", err, output)
	}
	root := filepath.Join(parent, "kb")
	if err := kb.Init(root); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, ".git")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("nested git repo: %v", err)
	}
}

func TestPreviewDoesNotWrite(t *testing.T) {
	root := filepath.Join(t.TempDir(), "absent")
	if err := kb.PreviewInit(root); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(root); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("dry run wrote root: %v", err)
	}
}

func TestInfo(t *testing.T) {
	root := t.TempDir()
	for _, note := range []string{"a.md", "nested/b.md", ".hidden/c.md", "nested/.cache/d.md", "other.txt"} {
		write(t, filepath.Join(root, note), "# Note")
	}
	if err := os.Symlink(filepath.Join(root, "a.md"), filepath.Join(root, "link.md")); err != nil {
		t.Fatal(err)
	}
	info, err := kb.Info(root)
	if err != nil || info.Root != root || info.Initialized || info.NoteCount != 2 {
		t.Fatalf("info = %+v, %v", info, err)
	}
	if _, err := kb.Info(filepath.Join(root, "missing")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing root: %v", err)
	}
}

func TestFixtureInfo(t *testing.T) {
	info, err := kb.Info("../../testdata/kb")
	if err != nil || info.NoteCount != 22 || info.Initialized {
		t.Fatalf("fixture info = %+v, %v", info, err)
	}
}

func TestFailedGitInitDoesNotMarkInitialized(t *testing.T) {
	root := filepath.Join(t.TempDir(), "new")
	t.Setenv("PATH", t.TempDir())
	if err := kb.Init(root); !errors.Is(err, kb.ErrWriteRefused) {
		t.Fatalf("git failure: %v", err)
	}
	info, err := kb.Info(root)
	if err != nil || info.Initialized {
		t.Fatalf("failed init: %+v, %v", info, err)
	}
}

func TestInfoThroughRootSymlink(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "real")
	if err := kb.Init(root); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(root, "note.md"), "# Note")
	link := filepath.Join(parent, "link")
	if err := os.Symlink(root, link); err != nil {
		t.Fatal(err)
	}
	info, err := kb.Info(link)
	if err != nil || !info.Initialized || info.Root != link || info.NoteCount != 1 {
		t.Fatalf("symlink root info = %+v, %v", info, err)
	}
}
