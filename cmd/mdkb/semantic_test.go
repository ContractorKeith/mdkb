package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/ContractorKeith/mdkb/internal/embed/embedtest"
)

func TestSemanticCLI(t *testing.T) {
	fixture(t)
	server := embedtest.Server(t)
	t.Setenv("MDKB_SEMANTIC", "1")
	t.Setenv("MDKB_OLLAMA_URL", server.URL)
	t.Setenv("MDKB_EMBED_MODEL", "test")
	if got := mustRun(t, "search", "shared", "--json"); !strings.Contains(got, `"mode":"keyword"`) {
		t.Fatal(got)
	}
	if _, err := run("search", "x", "--semantic"); exitCode(err) != 4 {
		t.Fatal(err)
	}
	dry := mustRun(t, "embed", "--dry-run", "--json")
	if !strings.Contains(dry, `"dry_run":true`) {
		t.Fatal(dry)
	}
	got := mustRun(t, "embed", "--json")
	if !strings.Contains(got, `"notes_embedded":22`) {
		t.Fatal(got)
	}
	got = mustRun(t, "embed", "--json")
	if !strings.Contains(got, `"skipped":22`) {
		t.Fatal(got)
	}
	got = mustRun(t, "search", "espresso", "--semantic", "--json")
	if !strings.Contains(got, `"mode":"semantic"`) || !strings.Contains(got, "spaces and unicode") {
		t.Fatal(got)
	}
	got = mustRun(t, "eval", "--json")
	var evaluation struct{ Results []any }
	if err := json.Unmarshal([]byte(got), &evaluation); err != nil || len(evaluation.Results) != 3 {
		t.Fatal(got, err)
	}
	server.Close()
	cmd := newRoot()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"search", "shared", "--json"})
	if err := cmd.Execute(); err != nil || !strings.Contains(stdout.String(), `"mode":"keyword"`) || !strings.Contains(stderr.String(), "keyword only") {
		t.Fatal(stdout.String(), stderr.String(), err)
	}
	if _, err := run("search", "x", "--semantic"); exitCode(err) != 4 {
		t.Fatal(err)
	}
	if _, err := run("embed", "--force"); exitCode(err) != 4 {
		t.Fatal(err)
	}
	mustRun(t, "reindex", "--full")
	got = mustRun(t, "info", "--json")
	if !strings.Contains(got, `"chunks":0`) {
		t.Fatal(got)
	}
	if _, err := run("eval", "--mode", "invalid"); err == nil {
		t.Fatal("accepted invalid mode")
	}
}

// A stopped fake server exercises the no-partial-output CLI failure path.
func TestEmbedAbsent(t *testing.T) {
	fixture(t)
	server := embedtest.Server(t)
	address := server.URL
	server.Close()
	t.Setenv("MDKB_OLLAMA_URL", address)
	out, err := run("embed")
	if exitCode(err) != 4 || out != "" || !strings.Contains(err.Error(), address) {
		t.Fatal(out, err)
	}
}
