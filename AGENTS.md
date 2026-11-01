# mdkb

A local-first knowledge base over a folder of plain Markdown files: a single Go binary with a
CLI and a stdio MCP server, keyword search (SQLite FTS5), opt-in semantic search (Ollama),
Obsidian-style wikilinks with backlinks, orphan and broken-link reports, and git-committed
agent writes. Files are the source of truth; the index is a rebuildable cache. It complements
an Obsidian vault and does not replace it. "Done" for v1 is every row marked v1 in
`docs/PLAN.md` §3 passing the acceptance checks in §8 on macOS and Linux.

## Status
M1 done — 2026-11-01: frontmatter, incremental FTS5 index, keyword search, and JSON contracts. M2–M6 remain.

## Commands
```bash
go build ./cmd/mdkb && ./mdkb version
go test ./...
test -z "$(gofmt -l .)" && go vet ./... && staticcheck ./...
```

## Pointers
- `docs/PLAN.md` — goal, feature matrix, architecture, CLI contract, milestones, acceptance.
- Session logs: `~/Projects/worklog/projects/mdkb.md` (not in this repo).

## Conventions & Gotchas
- `internal/kb` is the only package that touches the KB folder. `cmd/` and `internal/mcpserver`
  contain no file or SQL access.
- `internal/index` alone imports `database/sql`; `internal/search` owns ranked `Hit` results.
- Default KB root is `~/kb`. Never point tests or examples at `~/vault`.
- Every read command has `--json`; every write has `--dry-run`. Keep `docs/json.md` in step.
- Pure Go only (`modernc.org/sqlite`, no CGO) so cross-compiles stay one command.
- One milestone = one issue = one branch = one PR = one T3 thread in its own worktree.
- Commits: `<type>: <description>`, author @ContractorKeith only.

## Out of Scope
- TUI, web UI, themes. Obsidian is the UI.
- Cloud sync, self-update. git syncs, Homebrew updates.
- LLM summarize/organize features. The calling agent does that.
- Reading or writing `~/vault` before `import obsidian` lands (post-v1).
