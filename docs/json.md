# JSON and CLI contract

Global flags: --root, --json, --quiet. Root precedence is flag, MDKB_ROOT,
~/.config/mdkb/config.toml (root = "..."), then ~/kb. Roots become absolute;
a leading ~/ expands to the user's home. An explicit init path takes precedence
over all root configuration. Unset and empty settings fall through.

Successful --json output is one object and a newline on stdout:

- version: {"version":"dev"} (release builds may inject a version).
- info: {"links":0,"broken_links":0,"orphans":0,"problems":0,"indexed_notes":0,"index_path":"/absolute/cache/mdkb/<hash>.db","root":"/absolute/path","note_count":0,"initialized":true}.
  Counts regular *.md files recursively, excluding dot-directories and child symlinks; a symlink root is supported.
  An existing directory without the marker returns initialized:false with exit 0;
  uninitialized info is an inspector and never creates files (`indexed_notes:0`,
  `index_path:""`). Initialized info synchronizes the index. A missing root exits 2.
- init [path]: {"root":"/absolute/path","dry_run":false}.
  --dry-run validates without writing and returns dry_run:true. This shape does
  not assert that a dry run has initialized the directory.

--quiet suppresses human success output; explicit --json still emits data.
Errors are one line on stderr, with no success JSON or usage dump.

Exit codes follow PLAN §6: 0 success, 1 usage/config/runtime error, 2 not found,
3 KB not initialized, 4 index error, 5 write refused. info can inspect an
uninitialized directory. Other reads and reindex require an initialized KB
(exit 3 for missing or uninitialized roots); index open/query failures exit 4.

init creates 00-Inbox/, .gitignore, and the .mdkb/config.toml marker (version = 1).
It initializes git only if the root is not already inside a repository.
A regular marker file identifies an existing KB; repeated init leaves it
untouched. Non-empty directories without that marker and existing file paths
are refused with exit 5. No git commit is created by M0 init.

## M1 commands

All initialized reads synchronize the cache incrementally before answering.
The cache is disposable; `reindex --full` drops and rebuilds its tables. Its
location is `$XDG_CACHE_HOME/mdkb/<sha256(absolute-root)[:16]>.db`, falling back
to `~/.cache/mdkb/`. Rebuilding reproduces query output; database bytes and file
metadata are not a portable contract.

- `ls [--tag T] [--limit N]`:
  `{"notes":[{"id":"welcome","title":"Welcome","tags":["start"]}],"total":1,"truncated":false}`.
  IDs sort ascending. Total counts all notes matching the exact tag filter before
  limiting; truncated is true when more match than are returned. Default limit 50.
- `tags`: `{"tags":[{"name":"start","count":1}]}`.
  Counts distinct notes per tag; sort by count descending, then name ascending.
- `search <q> [--limit N]`:
  `{"hits":[{"id":"welcome","title":"Welcome","score":-0.5,"snippet":"[Welcome]"}]}`.
  Default limit 50. FTS5 BM25 scores rank ascending (lower is better), then ID.
  Public scores are rounded to 4 decimal places in `internal/search`, in both
  JSON and text output; JSON numbers may omit trailing zeroes. Ranking uses
  the original scores before rounding.
  Snippets bracket matches with `[` and `]`, using ` … ` for omissions.
  Invalid FTS expressions retry as a literal quoted phrase; no matches succeeds.
- `cat <id>`:
  `{"id":"welcome","title":"Welcome","tags":["start"],"aliases":["Home"],"body":"# Welcome\n"}`.
  Optional `problem` describes malformed frontmatter when present.
  JSON body excludes YAML frontmatter. Human output is the exact raw file,
  including frontmatter. Missing IDs exit 2; IDs omit `.md` and use `/` separators.
  Absolute/traversal IDs and child symlinks (including directory symlinks) are refused.
- `reindex [--full]`: `{"indexed_notes":22}` with `--json`.

Empty collections are `[]`, never null. Limits must be positive integers.
Frontmatter supports title, tags (a list or one string), aliases, created and
updated. Title falls back to the first H1 outside code fences, then the filename.
Invalid or unterminated frontmatter falls back to the raw file as the body,
with title from H1 or filename and empty metadata. Notes remain readable and
searchable. `ls --json` includes `problem` only on affected notes; `info` includes
`problems`, the number of affected notes. Filesystem errors still fail reads.

## M2 commands

All graph reads auto-sync. Link lines are one-based within the Markdown body,
excluding successfully parsed frontmatter. Fenced code (backtick or tilde, using
M1's title-fallback fence rule) and matched inline backtick spans are ignored.
Links and embeds support optional heading and display alias together.
Targets, headings and display aliases are trimmed; empty targets are ignored.

Resolution stops at the first hit: exact note ID (with or without `.md`, using
forward slashes), unique basename for a target without a folder, then unique
case-insensitive frontmatter alias. Exact IDs and basenames are case-sensitive.
Multiple basename matches are ambiguous and stop resolution before aliases;
multiple alias matches are also ambiguous. Explicit paths do not fall back to
a different folder's basename. Heading anchors are retained, not validated.
Graph changes reconsider existing matches, so new ambiguity, removed targets and
changed aliases are reflected without a full rebuild. M1 caches upgrade on sync.

- `links <id>`:
  `{"id":"welcome","links":[{"target":"plain","target_id":"plain","heading":"","alias":"","embed":false,"line":2,"resolved":true}]}`.
  Links stay in body order. `target` is the trimmed target without the heading
  or display alias; unresolved links have `target_id:null` and `resolved:false`.
  Text prints the resolved ID or `(broken) target`, one per line.
- `backlinks <id>`:
  `{"id":"plain","backlinks":[{"from":"welcome","line":2,"alias":"","embed":false}]}`.
  Includes each incoming occurrence, sorted by source ID, then line and body
  order. Text prints source ID, line and display alias. Both ID commands require
  an exact indexed ID (without `.md`); an unknown ID exits 2.
- `orphans`:
  `{"orphans":[{"id":"orphan","title":"Orphan"}],"total":1,"truncated":false}`.
  Notes without incoming resolved links, sorted by ID. Embeds and self-links
  count as incoming links.
- `broken`:
  `{"broken":[{"from":"broken-source","target":"does-not-exist","line":2,"reason":"missing"}],"total":1,"truncated":false}`.
  Unresolved occurrences sorted by source ID, then line and body order.
  Reason is `missing` or `ambiguous`. Text prints source, line, target and reason.
  Orphans and broken reports cap at 50 rows, with total counting all matches.
- `info` additionally includes `links`, `broken_links` and `orphans` integer
  totals, independent of report caps. These are zero for uninitialized roots.

All empty collections remain `[]`. Index rows store `kind` as `link` or `embed`,
with nullable `target_id`; display aliases never influence resolution.


## M3 MCP tools

See [MCP setup](mcp.md). Read tools use the same shapes above, with an optional
`problem` on `read_note` as on `cat`. MCP list/search limits default to 50 and
are capped at 50; non-positive limits are errors. Unknown IDs and refused writes
return `isError:true` tool results.

Both `write_note` and `capture` return
`{"id":"00-Inbox/example","path":"00-Inbox/example.md","bytes":42,"mode":"create","commit":true}`.
Path is relative to the root; bytes counts the entire resulting file.
Modes are create, append and overwrite; capture always uses create.
Neither write tool is registered without `--allow-write`.

## M4 write commands

All write IDs follow the MCP identifier rules. Refused writes (including invalid
IDs and uninitialized roots) exit 5; a missing source for `mv` exits 2.
`--json` produces one object on stdout; previews and commit diagnostics use stderr.

- `new <title> [--body TEXT | --stdin] [--tags a,b] [--dir DIR]`:
  `{"id":"meeting-notes","path":"meeting-notes.md","bytes":90,"mode":"create","commit":true}`.
  The ID uses the capture slug rules (Unicode letters/digits, lowercase, dashes,
  at most 40 characters, fallback `note`). Default directory is the root.
  Frontmatter includes title, optional tags and created (YYYY-MM-DD).
- `capture [--title T] [TEXT... | --stdin]`: the same WriteResult shape;
  creates `00-Inbox/YYYY-MM-DD-HHMMSS-slug.md` with the shared MCP capture rules.
- `daily [--append TEXT | --stdin] [--date YYYY-MM-DD]`: creates
  `10-Daily/YYYY-MM-DD.md` with created frontmatter and a date heading, or
  appends with a blank line. Creation/appending returns WriteResult; opening
  an existing note without text returns
  `{"id":"10-Daily/2026-11-01","path":"10-Daily/2026-11-01.md"}` without a write.
- `new`, `capture`, and writing `daily` accept `--dry-run`, returning
  `{"id":"meeting-notes","path":"meeting-notes.md","dry_run":true}`
  and the proposed complete content on stderr. They create no files/directories.
- `mv <old> <new>`:
  `{"old":"plain","new":"renamed/plain","rewritten":[{"id":"welcome","count":1}],"total":1,"commit":true}`.
  Rewritten entries count occurrences by original source ID. The actual fixture
  rewrites four occurrences across four sources. Existing destinations are refused.
  `--dry-run` adds `"dry_run":true`, sets commit:false and emits whole-file
  unified diffs on stderr, including the moved file. It writes no notes.
  Exact IDs (including `.md` spellings) become the new extensionless ID.
  A basename spelling remains only if it still resolves unambiguously to the
  moved note after the move; otherwise it becomes the new exact ID. Frontmatter
  aliases remain when they still resolve. Heading/display suffixes and embeds
  are preserved; links inside fenced or matched inline code are untouched.

### Write configuration and commits

User config is `~/.config/mdkb/config.toml`:

```toml
[write]
commit_on_write = true
daily_dir = "10-Daily"
```

`MDKB_COMMIT_ON_WRITE=0` or `1` overrides the boolean. A write inside a git work
tree stages and commits only its affected paths, once for the complete operation.
Messages use `mdkb: new <id>`, `mdkb: append <id>`,
`mdkb: overwrite <id>`, `mdkb: capture <id>`, `mdkb: daily <id>`, or
`mdkb: mv <old> -> <new> (N links rewritten)`.
If no git email is configured, the commit uses mdkb / mdkb@local for that command.

`commit` is true only when HEAD advanced. Disabled commits, non-repositories,
identical overwrites and changes restoring the committed content return false.
Commit failure preserves the write, prints a diagnostic to stderr, returns
commit:false and exits 0. Reindex failures remain errors after the file write.
The application layer owns reindex/commit orchestration to avoid a kb/index
dependency cycle; low-level `kb.Write` handles atomic file publication.
