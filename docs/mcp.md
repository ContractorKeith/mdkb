# MCP stdio server

Build with `go build ./cmd/mdkb`, initialize a separate KB, and use absolute
paths for the binary and root. Stdout belongs to MCP; diagnostics go to stderr.
An uninitialized or missing root exits 3 before transport startup.

Add it to Claude Code from the project where you want to use it:

```sh
claude mcp add --scope local mdkb -- /absolute/path/mdkb --root /absolute/path/my-kb mcp
claude mcp list
claude mcp remove --scope local mdkb
```

For Codex, add this to a trusted project's `.codex/config.toml` (or to
`~/.codex/config.toml` for all projects):

```toml
[mcp_servers.mdkb]
command = "/absolute/path/mdkb"
args = ["--root", "/absolute/path/my-kb", "mcp"]
```

See the [official Codex MCP configuration documentation](https://learn.chatgpt.com/docs/extend/mcp).

## Tools

| Tool | Input | Result |
|---|---|---|
| `list_notes` | `{tag?, limit?}` | `ls`: notes, total, truncated; at most 50 |
| `read_note` | `{id}` | `cat`: id, title, tags, aliases, body, optional problem |
| `search_notes` | `{query, limit?, mode?}` | `search`: mode and ranked hits; at most 50 |
| `get_links` | `{id}` | `links`: all outgoing occurrences in body order |
| `get_backlinks` | `{id}` | `backlinks`: all incoming occurrences in source order |
| `get_info` | `{}` | `info`: root, cache path, note and graph totals, embeddings {model, notes, chunks} |
| `write_note` | `{id, body, mode, frontmatter?}` | id, path, bytes, mode, commit |
| `capture` | `{text, title?}` | id, path, bytes, mode, commit |

Read outputs use the [CLI JSON contracts](json.md). Empty collections are arrays.
Limits default to 50, must be positive, and are capped at 50. Every call syncs
the index first. Unknown IDs, refused writes and storage failures return MCP
tool errors (`isError: true`) so the agent can correct them. These are errors
inside a registered tool. Unknown or unregistered tools return JSON-RPC
`-32602` (SDK behaviour), including write tools when `--allow-write` is absent.
One malformed JSON line ends the stdio session (SDK behaviour); the client
restarts the server.

IDs are paths relative to the KB root without `.md`, with `/` separators,
such as `projects/design`. Absolute paths, empty or dot segments, hidden path
components, traversal and child symlinks are refused. A symlink root is supported.

## Enable writes explicitly

Without `--allow-write`, neither write tool is registered or callable.
To enable them, append `--allow-write` after `mcp` in either client configuration.

`write_note` requires one of these modes:

- `create`: fail if the note already exists.
- `append`: require an existing note, insert a blank line, then add the body.
- `overwrite`: require an existing note and replace its contents.

Optional frontmatter accepts `title`, `tags` and `aliases` for create or overwrite;
append refuses frontmatter. YAML is serialized safely. Omitted frontmatter leaves
the supplied body unchanged apart from adding a missing final newline. All write
modes, including capture, ensure the file ends with a newline. `bytes` is the
resulting file size; `path` is relative to the root and includes `.md`.

Writes stage a complete temp file in the same directory. Append and overwrite
publish with rename; create uses an atomic hard link to refuse concurrent creates
without clobbering. Temporary files are removed afterward. The changed note is
reindexed, including graph resolution. A reindex failure reports that the file
was written; the next sync repairs the disposable cache.

`capture` creates `00-Inbox/YYYY-MM-DD-HHMMSS-slug.md` using local time, with
`created` frontmatter and an optional title. The slug uses the title or first
words of text, lowercase letters/digits and dashes, capped at 40 characters
(fallback: `note`). A same-second ID collision returns an error without replacing
the existing capture.

Both write tools use the shared CLI write path and create one git commit when
enabled and inside a repository. `commit` is true only when HEAD advances;
disabled commits, non-repositories and unchanged content return false.
A git failure preserves the write and returns a successful tool result with
`commit:false`, with a diagnostic on server stderr. See [write configuration](json.md#write-configuration-and-commits).
Create messages use `mdkb: new <id>`; capture messages use `mdkb: capture <id>`.


Search mode accepts `keyword`, `semantic` or `hybrid`. Omission selects hybrid
only when semantic is enabled and vectors exist for the configured model.
Otherwise it uses keyword. Explicit semantic/hybrid returns a tool error if
vectors are missing or Ollama is unavailable. Automatic hybrid falls back to
keyword with one diagnostic on server stderr. Hybrid hits include `sources`.
Configure embeddings with [semantic settings](json.md#semantic-search-m5) and
run `mdkb embed` through the CLI before semantic retrieval.
