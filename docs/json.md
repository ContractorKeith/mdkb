# JSON and CLI contract

Global flags: --root, --json, --quiet. Root precedence is flag, MDKB_ROOT,
~/.config/mdkb/config.toml (root = "..."), then ~/kb. Roots become absolute;
a leading ~/ expands to the user's home. An explicit init path takes precedence
over all root configuration. Unset and empty settings fall through.

Successful --json output is one object and a newline on stdout:

- version: {"version":"dev"} (release builds may inject a version).
- info: {"problems":0,"indexed_notes":0,"index_path":"/absolute/cache/mdkb/<hash>.db","root":"/absolute/path","note_count":0,"initialized":true}.
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
Wikilinks remain plain text in M1.
