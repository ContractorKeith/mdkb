# mdkb

Local-first knowledge base over plain Markdown files. One Go binary: CLI plus MCP server,
FTS5 keyword search, opt-in local embeddings, Obsidian-style wikilinks and backlinks,
git-committed agent writes. Your files stay the source of truth.

**Status:** M0 scaffold complete (2026-11-01). Search, links, and MCP are planned
for later milestones; see [the plan](docs/PLAN.md).

Requires Go 1.27 and git on PATH.

```bash
go build ./cmd/mdkb
./mdkb version
./mdkb init /tmp/example-kb --dry-run
./mdkb init /tmp/example-kb
./mdkb --root /tmp/example-kb info --json
```

Root precedence: `--root` > `MDKB_ROOT` > `~/.config/mdkb/config.toml`
(`root = "..."`) > `~/kb`. An explicit `init [path]` wins over configuration.
See [CLI output contracts](docs/json.md) and [fixture notes](testdata/kb/README.md).
