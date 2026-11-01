# M0 JSON and CLI contract

Global flags: --root, --json, --quiet. Root precedence is flag, MDKB_ROOT,
~/.config/mdkb/config.toml (root = "..."), then ~/kb. Roots become absolute;
a leading ~/ expands to the user's home. An explicit init path takes precedence
over all root configuration. Unset and empty settings fall through.

Successful --json output is one object and a newline on stdout:

- version: {"version":"dev"} (release builds may inject a version).
- info: {"root":"/absolute/path","note_count":0,"initialized":true}.
  Counts regular *.md files recursively, excluding dot-directories and child symlinks; a symlink root is supported.
  An existing directory without the marker returns initialized:false with exit 0;
  info is an inspector and never creates files. A missing root exits 2.
- init [path]: {"root":"/absolute/path","dry_run":false}.
  --dry-run validates without writing and returns dry_run:true. This shape does
  not assert that a dry run has initialized the directory.

--quiet suppresses human success output; explicit --json still emits data.
Errors are one line on stderr, with no success JSON or usage dump.

Exit codes follow PLAN §6: 0 success, 1 usage/config/runtime error, 2 not found,
3 KB not initialized, 4 index error, 5 write refused. M0 info can inspect an
uninitialized directory, and M0 has no index; codes 3 and 4 are reserved for
commands requiring those facilities.

init creates 00-Inbox/, .gitignore, and the .mdkb/config.toml marker (version = 1).
It initializes git only if the root is not already inside a repository.
A regular marker file identifies an existing KB; repeated init leaves it
untouched. Non-empty directories without that marker and existing file paths
are refused with exit 5. No git commit is created by M0 init.
