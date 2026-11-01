package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func cacheInfo(t *testing.T, args ...string) map[string]any {
	t.Helper()
	var info map[string]any
	out, err := run(append([]string{"info", "--json"}, args...)...)
	if err != nil {
		t.Errorf("info failed: %v", err)
		return nil
	}
	if err := json.Unmarshal([]byte(out), &info); err != nil {
		t.Fatal(err)
	}
	return info
}

func TestCorruptCacheRecovery(t *testing.T) {
	writeFixture(t)
	before := mustRun(t, "ls", "--json")
	path := cacheInfo(t)["index_path"].(string)
	if err := os.WriteFile(path, []byte(strings.Repeat("not sqlite!", 400)), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := run("ls")
	if exitCode(err) != 4 || !strings.Contains(err.Error(), path) || !strings.Contains(err.Error(), "run mdkb reindex --full") {
		t.Errorf("opaque cache error: %v", err)
	}
	info := cacheInfo(t)
	if info["index_path"] != path || info["indexed_notes"] != float64(0) {
		t.Errorf("corrupt info: %+v", info)
	}
	problem, ok := info["problem"].(map[string]any)
	if !ok || problem["index"] != "unreadable" {
		t.Errorf("missing cache problem: %+v", info)
	}
	for _, suffix := range []string{"-wal", "-shm"} {
		if err := os.WriteFile(path+suffix, []byte("stale sidecar"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := run("reindex", "--full", "--json"); err != nil {
		t.Fatalf("full reindex cannot recover: %v", err)
	}
	if after := mustRun(t, "ls", "--json"); after != before {
		t.Fatalf("recovery changed listing: %s", after)
	}
	for _, suffix := range []string{"-wal", "-shm"} {
		raw, err := os.ReadFile(path + suffix)
		if err == nil && string(raw) == "stale sidecar" {
			t.Errorf("stale sidecar retained: %s", suffix)
		}
	}
}

func TestCacheRootAliases(t *testing.T) {
	root := writeFixture(t)
	before := cacheInfo(t)["index_path"]
	link := filepath.Join(t.TempDir(), "alias")
	if err := os.Symlink(root, link); err != nil {
		t.Fatal(err)
	}
	if after := cacheInfo(t, "--root", link)["index_path"]; after != before {
		t.Fatalf("root aliases use different caches: %v != %v", after, before)
	}
}
