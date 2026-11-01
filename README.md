# mdkb

Local-first knowledge base over plain Markdown files. One Go binary: CLI plus MCP server,
FTS5 keyword search, opt-in local embeddings, Obsidian-style wikilinks and backlinks,
git-committed agent writes. Your files stay the source of truth.

**Status:** M3 complete (2026-11-01): keyword search, links, and MCP stdio tools.
CLI writes, embeddings, and release packaging remain; see [the plan](docs/PLAN.md).

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
See [MCP setup and tools](docs/mcp.md), [CLI output contracts](docs/json.md) and [fixture notes](testdata/kb/README.md).
