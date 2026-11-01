# mdkb

Local-first knowledge base over plain Markdown files. One Go binary: CLI plus MCP server,
FTS5 keyword search, opt-in local embeddings, Obsidian-style wikilinks and backlinks,
git-committed agent writes. Your files stay the source of truth.

**Status:** Skeptic round 1 fixes awaiting orchestrator review (2026-11-01).
Mac Studio real-model evaluation, tag and tap publication remain.
See [the plan](docs/PLAN.md).

## Install

After v0.1.0 and its tap formula are published:

```bash
brew install contractorkeith/tap/mdkb
```

From source (Go 1.27 required):

```bash
go install github.com/ContractorKeith/mdkb/cmd/mdkb@latest
export PATH="$(go env GOPATH)/bin:$PATH"
```

Source installs report `mdkb dev`; release binaries report the release tag.

Or download the matching macOS/Linux amd64/arm64 tarball from
[GitHub Releases](https://github.com/ContractorKeith/mdkb/releases), verify it
against `checksums.txt` (`sha256sum` on Linux, `shasum -a 256` on macOS), extract
it, and place `mdkb` on your PATH. Git on PATH enables commits for writes.

macOS tarball binaries are unsigned, so Gatekeeper may block them. Prefer
Homebrew or `go install`. After verifying the checksum, you can remove the
download quarantine with `xattr -d com.apple.quarantine mdkb`.

## Quick start (60 seconds)

```bash
# A new, separate scratch KB; explicit roots keep existing folders untouched.
example_root="$(mktemp -d)/mdkb-example"
mdkb init "$example_root"
mdkb --root "$example_root" new "Welcome" --body 'A local Markdown knowledge base.'
mdkb --root "$example_root" new "Ideas" --body 'Start with [[welcome]].'
mdkb --root "$example_root" search 'Markdown'
mdkb --root "$example_root" backlinks welcome
# Register the read-only stdio MCP server in Claude Code:
claude mcp add --scope local mdkb -- "$(command -v mdkb)" --root "$example_root" mcp
claude mcp list
```

Ask Claude Code to list or search your notes. Add `--allow-write` after `mcp`
to enable its two write tools. Every read supports `--json`; CLI writes support
`--dry-run`. Optional semantic search uses local Ollama; see
[semantic settings](docs/json.md#semantic-search-m5) and [evaluation](docs/search-eval.md).

## Obsidian relationship

mdkb complements Obsidian: use Obsidian as the editor and mdkb for CLI/agent access.
It owns a separate Markdown folder (default `~/kb`). Your existing vault stays
untouched. An `import obsidian` command is planned after v1; there is no vault
import in this release.

Root precedence: `--root` > `MDKB_ROOT` > `~/.config/mdkb/config.toml`
(`root = "..."`) > `~/kb`. An explicit `init [path]` wins over configuration.
See [MCP setup and tools](docs/mcp.md), [CLI output contracts](docs/json.md),
[fixture notes](testdata/kb/README.md) and [Homebrew publishing](packaging/homebrew/README.md).
