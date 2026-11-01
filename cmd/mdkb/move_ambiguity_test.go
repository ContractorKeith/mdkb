package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMovePreservesOtherTargets(t *testing.T) {
	for _, tc := range []struct{ name, existing, destination, body string }{
		{"ambiguity", "folder/shared", "other/shared", "![[shared#Heading|Alias]]\n"},
		{"retarget", "other/b", "b", "![[b#Heading|Alias]]\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			t.Setenv("HOME", t.TempDir())
			t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(t.TempDir(), "gitconfig"))
			t.Setenv("XDG_CACHE_HOME", t.TempDir())
			t.Setenv("MDKB_ROOT", root)
			mustRun(t, "init", root)
			t.Setenv("MDKB_COMMIT_ON_WRITE", "0")
			for id, body := range map[string]string{"a": "# A\n", tc.existing: "# Existing\n", "source": tc.body} {
				name := filepath.Join(root, id+".md")
				if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(name, []byte(body), 0644); err != nil {
					t.Fatal(err)
				}
			}
			out := mustRun(t, "mv", "a", tc.destination, "--json")
			source, err := os.ReadFile(filepath.Join(root, "source.md"))
			if err != nil {
				t.Fatal(err)
			}
			got := out + string(source)
			want, err := os.ReadFile(filepath.Join("../../testdata/golden", "mv-"+tc.name+".txt"))
			if err != nil || got != string(want) {
				t.Fatalf("golden mismatch: %v\n%s", err, got)
			}
			links := mustRun(t, "links", "source", "--json")
			if !strings.Contains(links, `"target_id":"`+tc.existing+`"`) {
				t.Fatalf("target changed: %s", links)
			}
		})
	}
}
