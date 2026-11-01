package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ContractorKeith/mdkb/internal/kb"
)

func run(args ...string) (string, error) {
	cmd := newRoot()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func TestCommandContract(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("MDKB_ROOT", "")
	root := filepath.Join(home, "kb")
	out, err := run("init", root)
	if err != nil || !strings.Contains(out, root) {
		t.Fatalf("init: %q, %v", out, err)
	}
	out, err = run("--root", root, "info", "--json", "--quiet")
	var info kb.Details
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(out), &info); err != nil {
		t.Fatal(err)
	}
	if info.Root != root || !info.Initialized || info.NoteCount != 0 {
		t.Fatalf("info: %+v", info)
	}
	for _, tt := range []struct {
		args   []string
		code   int
		output bool
	}{
		{[]string{"version"}, 0, true},
		{[]string{"version", "--json"}, 0, true},
		{[]string{"version", "--quiet"}, 0, false},
		{[]string{"--version"}, 0, true},
		{[]string{"--help"}, 0, true},
		{[]string{"init", root, "--quiet"}, 0, false},
		{[]string{"info", "--root", filepath.Join(home, "missing")}, 2, false},
		{[]string{"init", home}, 5, false},
		{[]string{"unknown"}, 1, false},
		{[]string{"info", "extra"}, 1, false},
		{[]string{"init", "a", "b"}, 1, false},
		{[]string{"info", "--bad"}, 1, false},
	} {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			out, err := run(tt.args...)
			if code := exitCode(err); code != tt.code {
				t.Fatalf("code %d, want %d: %v", code, tt.code, err)
			}
			if (out != "") != tt.output {
				t.Fatalf("unexpected output %q", out)
			}
		})
	}
	dry := filepath.Join(home, "dry")
	if _, err := run("init", dry, "--dry-run"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(dry); !os.IsNotExist(err) {
		t.Fatalf("dry run: %v", err)
	}
}
