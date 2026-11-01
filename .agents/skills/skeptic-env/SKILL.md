---
name: skeptic-env
description: "How a skeptical reviewer builds, seeds and attacks mdkb. Read by skeptic-review; not a workflow on its own."
---

# skeptic-env — mdkb

Go CLI + stdio MCP over Markdown; files are truth, SQLite is disposable cache.

## Stack and commands

```bash
go build -o /tmp/mdkb-skeptic ./cmd/mdkb
go test -race ./...
test -z "$(gofmt -l .)" && go vet ./... && staticcheck ./...
# Repeat builds with GOOS=linux/darwin and GOARCH=amd64/arm64:
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o /tmp/mdkb-cross ./cmd/mdkb
# Build/test before isolating HOME (which also moves Go's caches).
scratch=$(mktemp -d)
export HOME="$scratch/home" XDG_CACHE_HOME="$scratch/cache"
export XDG_CONFIG_HOME="$scratch/config" GIT_CONFIG_GLOBAL="$scratch/gitconfig"
export GIT_CONFIG_NOSYSTEM=1 MDKB_ROOT="$scratch/kb"
mkdir -p "$HOME" "$XDG_CACHE_HOME" "$XDG_CONFIG_HOME"
unset MDKB_COMMIT_ON_WRITE MDKB_SEMANTIC MDKB_EMBED_MODEL MDKB_OLLAMA_URL
/tmp/mdkb-skeptic init "$MDKB_ROOT"
cp -R testdata/kb/. "$MDKB_ROOT/"
/tmp/mdkb-skeptic --root "$MDKB_ROOT" reindex
```

Never use the reviewer's KB, config or vault. All CLI calls use the scratch root.

## Platforms

| Platform | Ships | Review on |
|---|---|---|
| Linux amd64/arm64 | yes | Omarchy |
| macOS amd64/arm64 | yes | Mac Studio |
| Windows | no | out of scope |

## What to attack

- IDs/paths: traversal, absolute/hidden paths, symlinks, Unicode and collisions.
- Wikilinks: exact path, unique basename, aliases, ambiguity after add/delete;
  headings/embeds and `mv` inbound rewriting, dry-run and repeated moves.
- MCP over stdio: hold stdin open through initialize and tool responses; six
  reads, eight tools with `--allow-write`; errors and concurrent tool calls.
- Writes: create/append/overwrite safety, dry-run, newline/frontmatter, git
  commit scope, dirty index, disabled commits, missing identity and hook failure.
- Ollama absent/present/killed mid-embed; partial failure, stale vectors,
  fallback vs explicit semantic errors. Mac Studio real-model eval procedure:
  [docs/search-eval.md](../../../../docs/search-eval.md).
- Delete the scratch cache DB, `reindex --full`, compare results; upgrade an
  M1-shaped DB (no problem column/link tables), malformed frontmatter recovery.
- Large KB: generate 5,000 linked notes in scratch and time `ls`, `search`,
  `broken` after initial index and after changes. Reproduce the generator in
  `internal/index/large_test.go`; `go test ./internal/index -run Large -v`.
- Install: all four release tarballs, injected CLI/MCP version, checksums and
  Homebrew formula after placeholder replacement. No tag during a review.

## Traps

SDK v1.8 cancels pending requests at stdin EOF. Some sandboxes block `rm -rf`;
use temporary directories and targeted cleanup. Scores round to four decimal
places. `info` parses every note. Fixture broken count is three; `mv` must add
zero new broken links. Semantic real-model results remain pending Mac Studio.

## Earlier rounds

Pre-release orchestrator reviews on PRs #10/#12 requested cached resolution and
config read outside kb; M6a addresses both. No skeptic round completed yet.
