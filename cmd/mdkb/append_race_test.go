package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestAppendProcessRace(t *testing.T) {
	root, home := t.TempDir(), t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(home, "gitconfig"))
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	t.Setenv("MDKB_ROOT", root)
	t.Setenv("MDKB_COMMIT_ON_WRITE", "0")
	mustRun(t, "init", root)
	mustRun(t, "daily", "--date", "2026-11-01", "--append", "seed")
	type process struct {
		cmd    *exec.Cmd
		out    bytes.Buffer
		marker string
	}
	processes := make([]process, 20)
	for i := range processes {
		p := &processes[i]
		p.marker = fmt.Sprintf("marker-%02d", i)
		p.cmd = exec.Command(testBinary, "daily", "--date", "2026-11-01", "--append", p.marker, "--json")
		p.cmd.Stdout, p.cmd.Stderr = &p.out, &p.out
		if err := p.cmd.Start(); err != nil {
			t.Fatal(err)
		}
	}
	successes := []string{}
	for i := range processes {
		p := &processes[i]
		if err := p.cmd.Wait(); err != nil {
			t.Errorf("%s: %v: %s", p.marker, err, p.out.String())
		} else {
			successes = append(successes, p.marker)
		}
		if strings.Contains(p.out.String(), "SQLITE_BUSY") {
			t.Errorf("busy surfaced: %s", p.out.String())
		}
	}
	raw, err := os.ReadFile(filepath.Join(root, "10-Daily/2026-11-01.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, marker := range successes {
		if !strings.Contains(string(raw), "\n"+marker+"\n") {
			t.Errorf("acknowledged append lost: %s", marker)
		}
	}
	if len(successes) != 20 {
		t.Errorf("only %d/20 appends succeeded", len(successes))
	}
	if _, err := os.Stat(filepath.Join(root, ".mdkb/write.lock")); err != nil {
		t.Errorf("missing write lock: %v", err)
	}
	t.Setenv("MDKB_COMMIT_ON_WRITE", "1")
	mustRun(t, "new", "committed", "--body", "tracked")
	if tracked := gitKB(t, root, "ls-files"); strings.Contains(tracked, ".mdkb/") {
		t.Fatalf("lock/config committed: %s", tracked)
	}
	if ignored := gitKB(t, root, "check-ignore", ".mdkb/write.lock"); ignored != ".mdkb/write.lock" {
		t.Fatal(ignored)
	}
}
