# Changelog

## Unreleased

Nothing yet.

## 0.1.0 — 2026-11-01

- Keep a knowledge base in ordinary Markdown files, initialized with git;
  inspect notes, tags and KB information from one local CLI.
- Search by keyword with SQLite FTS5, snippets and a rebuildable incremental
  cache outside the KB. Read commands offer stable JSON and bounded listings.
- Follow Obsidian-style wikilinks, aliases, headings and embeds; inspect
  backlinks, orphans and broken links. Malformed frontmatter remains readable.
- Connect agents through a stdio MCP server with six read tools and two write
  tools enabled explicitly with `--allow-write`.
- Create, capture, append and overwrite notes; maintain daily notes and rename
  notes while rewriting inbound links. Preview CLI writes with `--dry-run`;
  scoped git commits preserve an undo history.
- Opt into local Ollama embeddings and semantic/hybrid search with reciprocal
  rank fusion. Keyword fallback keeps automatic search usable without Ollama;
  real-model ranking evaluation remains pending on Mac Studio.
- Prepare macOS/Linux release archives, checksums, a Homebrew formula template
  and isolated skeptic review instructions. v0.1.0 awaits skeptic review and tag.
- Fixed before tagging, from the skeptic round on both platforms (#15–#23): a
  cross-process write lock so parallel appends never lose data; indexing and
  reads confined to the KB root even under symlink swaps; `mv` keeps every
  other link pointing where it did; ids with wikilink delimiters are refused;
  every listed id is readable; a corrupt cache names its path and
  `reindex --full` recovers it; NFC/NFD link targets agree on APFS; an inline
  triple-backtick span no longer hides later links.
