# mdkb

A local-first knowledge base over a folder of plain Markdown files: a single Go binary with a
CLI and a stdio MCP server, keyword search (SQLite FTS5), opt-in semantic search (Ollama),
Obsidian-style wikilinks with backlinks, orphan and broken-link reports, and git-committed
agent writes. Files are the source of truth; the index is a rebuildable cache. It complements
an Obsidian vault and does not replace it. "Done" for v1 is every row marked v1 in
`docs/PLAN.md` §3 passing the acceptance checks in §8 on macOS and Linux.

## Status
v0.1.0 released — 2026-11-01: M0–M6 shipped, skeptic round 1 (#15–#23) fixed in #24, Homebrew tap published, real-model semantic eval done on the Mac Studio (hybrid 0.90/0.9067 vs keyword 0.70/0.70). Next: post-v1 backlog in docs/PLAN.md §3 Later rows.

## Commands
```bash
bash scripts/accept.sh          # the acceptance: gate, build, goldens, live checks; CI runs this
go build ./cmd/mdkb && ./mdkb version
go test ./...
test -z "$(gofmt -l .)" && go vet ./... && staticcheck ./...
```

## Pointers
- `docs/PLAN.md` — goal, feature matrix, architecture, CLI contract, milestones, acceptance.
- Session logs: `~/Projects/worklog/projects/mdkb.md` (not in this repo).

## Conventions & Gotchas
- `internal/kb` is the only package that touches the KB folder. `cmd/` and `internal/mcpserver`
  contain no file or SQL access; MCP tools share auto-sync through `internal/app`.
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
