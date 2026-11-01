package main

import (
	"os"
	"strings"
	"testing"
)

func TestInstallAndProtocolDocs(t *testing.T) {
	for file, fragments := range map[string][]string{
		"README.md":                           {`export PATH="$(go env GOPATH)/bin:$PATH"`, "mdkb dev", "Gatekeeper", "xattr -d com.apple.quarantine mdkb"},
		"docs/mcp.md":                         {"-32602", "unregistered", "isError", "malformed JSON", "restarts"},
		"docs/json.md":                        {"HTML comments", "known limitation"},
		"packaging/homebrew/mdkb.rb":          {"# frozen_string_literal: true", "# Local-first Markdown knowledge base."},
		"AGENTS.md":                           {"v0.1.0", "#15–#23"},
		".agents/skills/skeptic-env/SKILL.md": {"2026-11-01", "3c582fc", "worklog/reports/mdkb-review-2026-11-01.md", "#15–#23"},
	} {
		raw, err := os.ReadFile("../../" + file)
		if err != nil {
			t.Fatal(err)
		}
		for _, fragment := range fragments {
			if !strings.Contains(string(raw), fragment) {
				t.Errorf("%s missing %q", file, fragment)
			}
		}
	}
	raw, err := os.ReadFile("../../docs/PLAN.md")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "mdkb doctor") {
		t.Error("PLAN advertises unimplemented doctor")
	}
}
