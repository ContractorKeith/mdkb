# mdkb — Build Plan

Status: planned, no code yet (2026-11-01). Written after reading
[devjarus/personal-knowledge-base](https://github.com/devjarus/personal-knowledge-base),
[ILDaviz/lyra-tui](https://github.com/ILDaviz/lyra-tui) (archived 2026-10-13) and
[sazardev/shiki](https://github.com/sazardev/shiki).

## 1. Goal

A single binary, `mdkb`, that turns a folder of plain Markdown files into a knowledge base an
agent and a human can both use:

1. **Files are the truth.** Notes are `*.md` with optional YAML frontmatter. Everything else is a
   rebuildable cache. The folder must stay `cat`/`grep`/`git`/`rsync`-able and openable in
   Obsidian as-is.
2. **Search** by keyword (SQLite FTS5) from day one, and by meaning (local embeddings via
   Ollama) as an opt-in.
3. **Links** are Obsidian-style `[[wikilinks]]`, with backlinks, orphans and broken-link reports.
4. **Agents** reach it through a stdio MCP server; humans through the CLI. Both sit on one
   core package. Every read has `--json`; every write has `--dry-run` and a git commit.

v1 is done when every row marked **v1** in the feature matrix (§3) works on macOS and Linux
and passes the acceptance checks in §8.

### Relationship to the Obsidian vault

mdkb owns a **separate** KB folder (default `~/kb`). It does not read or write `~/vault`.
An `import obsidian` command (post-v1) will copy a vault in while keeping link semantics.
This keeps 1,454 real notes out of the blast radius while the tool is young.

## 2. What the inspirations got right (and what to skip)

| Source | Copy | Skip |
|---|---|---|
| personal-knowledge-base | one `core/` layer that alone touches the filesystem; index as sidecar; hybrid search | S3 sync, LLM clustering, Next.js viewer |
| lyra-tui | link model: aliases, heading anchors, embeds; orphan and unresolved reports; `--json`, `--dry-run` | TUI, bookmarks, git scheduler |
| shiki | git as the undo log for agent writes; rename rewrites inbound links; list caps so agents don't flood context | 37 themes, browser extension, PDF publish, self-updater |

## 3. Feature matrix

**v1** = in scope now. **Later** = after v1. **Drop** = deliberately never.

| # | Feature | Decision | Milestone | Notes |
|---|---|---|---|---|
| 1 | `init` creates KB folder, config, `.gitignore`, git repo | v1 | M0 | |
| 2 | Frontmatter parse: `title`, `tags`, `aliases`, `created`, `updated` | v1 | M1 | All optional; title falls back to H1 then filename |
| 3 | Incremental index (mtime+size) in SQLite outside the KB | v1 | M1 | `$XDG_CACHE_HOME/mdkb/<root-hash>.db` |
| 4 | `ls`, `cat`, `tags`, `info`, `reindex` | v1 | M1 | `ls` capped at 50 unless `--limit` |
| 5 | Keyword search, FTS5 bm25, snippets | v1 | M1 | `search <query>` |
| 6 | `--json` on every read command | v1 | M1 | Stable shapes documented in `docs/json.md` |
| 7 | Wikilinks: `[[note]]`, `[[note\|alias]]`, `[[note#heading]]`, `![[embed]]` | v1 | M2 | Resolution: exact path, then unique basename, then frontmatter alias |
| 8 | `links`, `backlinks`, `orphans`, `broken` | v1 | M2 | |
| 9 | MCP stdio server, read tools | v1 | M3 | `list_notes`, `read_note`, `search_notes`, `get_links`, `get_backlinks`, `get_info` |
| 10 | MCP write tools behind `--allow-write` | v1 | M3 | `write_note` (mode: create, append, overwrite), `capture` |
| 11 | `new`, `capture` (to `00-Inbox/`), `daily` | v1 | M4 | `--body`, `--stdin`, `--tags`, `--dry-run` |
| 12 | `mv` rewrites inbound links | v1 | M4 | `--dry-run` shows the diff |
| 13 | Git commit per write when KB is a repo | v1 | M4 | `commit_on_write` config, default on |
| 14 | Embeddings via Ollama, `embed`, hybrid search | v1 | M5 | Reciprocal rank fusion, not fixed weights; brute-force cosine is fine under ~20k notes |
| 15 | Cross-compile macOS/Linux, Homebrew tap, `v0.1.0` | v1 | M6 | Windows later |
| 16 | `import obsidian <vault> [--dry-run]` | Later | | After v1 proves link semantics on fixtures |
| 17 | Unlinked mentions | Later | | |
| 18 | `watch` mode (fsnotify) | Later | | `reindex` is cheap enough for now |
| 19 | MCP resources (`mdkb://note/<id>`) | Later | | Tools first; add when a client needs them |
| 20 | TUI, web UI, themes | Drop | | Obsidian is the UI |
| 21 | Cloud sync, self-updater | Drop | | git is sync; Homebrew is update |
| 22 | LLM organize/summarize | Drop | | The calling agent does that |

## 4. Architecture

```
cmd/mdkb/            cobra entrypoint only; no logic
internal/kb/         the ONLY package that touches the KB folder: walk, parse, links, write, git
internal/index/      SQLite schema, incremental sync, FTS5 queries, embeddings table
internal/search/     keyword, semantic, fusion; returns ranked []Hit
internal/mcpserver/  tool registration over internal/app + internal/kb + internal/search
internal/config/     ~/.config/mdkb/config.toml, env MDKB_ROOT, --root flag precedence
testdata/kb/         golden fixture KB (~20 notes) every package tests against
docs/PLAN.md         this file
docs/json.md         JSON output contracts (M1)
```

Rules:
- `cmd/` and `internal/mcpserver/` never import `database/sql` or touch files. They call
  `kb` and `search`.
- The index can be deleted at any time and `reindex --full` must rebuild it bit-for-bit.
- Note id = path relative to the KB root without `.md`. That is the only identifier.
- Writes go through `kb.Write*`, which does: dry-run check, write temp + rename, reindex that
  file, git add + commit if enabled.

### Dependencies and why

| Module | Why |
|---|---|
| `github.com/spf13/cobra` | Standard Go CLI; `create-cli` skill patterns apply |
| `modernc.org/sqlite` v1.61+ | Pure Go, FTS5 verified on 2026-11-01; no CGO keeps cross-compile trivial |
| `github.com/modelcontextprotocol/go-sdk/mcp` | Official SDK; typed tool handlers, stdio transport |
| `gopkg.in/yaml.v3` | Frontmatter |
| `github.com/BurntSushi/toml` | Config |

No goldmark or Markdown renderer. We only need frontmatter, headings and wikilinks, and a
small hand-written scanner is easier to test than a full AST.

## 5. Data model

**Note** `{id, path, title, aliases[], tags[], created, updated, mtime, size, body}`.
**Link** `{from_id, target_raw, target_id|null, kind: link|embed, heading, alias, line}`.
`target_id` null means broken.

SQLite tables: `notes`, `links`, `tags`, `notes_fts` (FTS5 external-content on `notes`),
`embeddings(note_id, model, chunk, vec BLOB)`.

## 6. CLI contract

```
mdkb init [path]                 mdkb search <q> [--semantic|--keyword] [--limit N] [--json]
mdkb ls [--tag T] [--limit N]    mdkb links <id>      mdkb backlinks <id>
mdkb cat <id>                    mdkb orphans         mdkb broken
mdkb tags                        mdkb new <title> [--body|--stdin] [--tags a,b] [--dry-run]
mdkb info                        mdkb capture [--title] [--stdin]
mdkb reindex [--full]            mdkb daily [--append TEXT]
mdkb embed [--model M]           mdkb mv <old> <new> [--dry-run]
mdkb mcp [--allow-write]         mdkb doctor
```

Exit codes: 0 ok, 1 usage, 2 not found, 3 KB not initialized, 4 index error, 5 write refused.
Global flags: `--root`, `--json`, `--quiet`. Config precedence: flag > env > config file > default.

## 7. Milestones

Each milestone is one GitHub issue, one branch, one PR, one T3 thread in its own worktree.

| M | Deliverable | Done when |
|---|---|---|
| M0 | Scaffold: cobra skeleton, `version`, `init`, config, CI (test, vet, gofmt, staticcheck), fixture KB | `mdkb init /tmp/x && mdkb info` works; CI green |
| M1 | Index and keyword search | Rows 2–6 pass §8 on the fixture |
| M2 | Links | Rows 7–8; every fixture link shape resolves or reports broken |
| M3 | MCP server | Rows 9–10; Claude Code lists and calls every tool; write refused without flag |
| M4 | Writes | Rows 11–13; `mv` on fixture leaves zero broken links; each write is one commit |
| M5 | Semantic search | Row 14; hybrid beats keyword on a 10-query fixture set, documented |
| M6 | Release | Row 15; `skeptic-review` findings triaged; `brew install contractorkeith/tap/mdkb` |

## 8. Acceptance checks

Run from a clean clone on macOS and Linux:

1. `go test ./...` green, `gofmt -l .` empty, `go vet ./...` and `staticcheck ./...` clean.
2. `mdkb init /tmp/kb && cp -r testdata/kb/* /tmp/kb && mdkb --root /tmp/kb reindex` then
   `search`, `ls --json`, `backlinks`, `orphans`, `broken` match the golden outputs in
   `testdata/golden/`.
3. Delete the index file, `reindex --full`, re-run step 2: identical output.
4. `mdkb mcp` under Claude Code: `list_notes` and `search_notes` return; `write_note` returns
   an error without `--allow-write` (tool absent); with it, the note exists and is
   reindexed. Git commits arrive in M4.
5. `mdkb mv` on a note with three inbound links: `broken` reports zero afterward; `--dry-run`
   changed nothing.
6. `ls` with 60 notes returns 50 and a `truncated: true` field in JSON.

## 9. Process (the real deliverable)

This project exists to tune the workflow as much as to ship the tool.

- **Per milestone**: one T3 thread in a worktree. `implementer` worker builds against the
  acceptance row with `tdd`. `reviewer` worker reads the full branch diff. `code-review` skill
  before the PR leaves draft. PR links and watches through T3.
- **Before tag**: `skeptic-review` on main; one labelled issue per confirmed bug.
- **After each session**: `log-work`. Skill gaps found go back to dotfiles, not patched here.
- **Tracking**: GitHub milestones M0–M6 and one issue each are the truth. The Basecamp
  "github" project card table (all personal repos share it) mirrors milestone state across all repos for the
  what's-next view; a card moves when a PR merges.
