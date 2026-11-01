#!/usr/bin/env bash
# scripts/accept.sh — the acceptance for mdkb (docs/PLAN.md §8). Worker, reviewer and CI
# run this exact script. Bash 3.2 compatible. Takes about a minute.
set -eu

cd "$(dirname "$0")/.."
scratch="$(mktemp -d)"
trap 'rm -r "$scratch"' EXIT
bin="$scratch/mdkb"

echo "== gate"
go test -race ./...
test -z "$(gofmt -l .)"
go vet ./...
staticcheck ./...

echo "== build"
go build -trimpath -ldflags "-s -w -X main.version=v0.0.0-accept" -o "$bin" ./cmd/mdkb
"$bin" version

echo "== fixture"
export HOME="$scratch/home" XDG_CACHE_HOME="$scratch/cache" XDG_CONFIG_HOME="$scratch/config"
export GIT_CONFIG_GLOBAL="$scratch/gitconfig" GIT_CONFIG_NOSYSTEM=1 MDKB_ROOT="$scratch/kb"
export MDKB_COMMIT_ON_WRITE=0
unset MDKB_SEMANTIC MDKB_EMBED_MODEL MDKB_OLLAMA_URL
mkdir -p "$HOME" "$XDG_CACHE_HOME" "$XDG_CONFIG_HOME"
"$bin" init "$MDKB_ROOT" >/dev/null
cp -R testdata/kb/. "$MDKB_ROOT/"

echo "== goldens"
golden() { # name, args...
  name="$1"; shift
  "$bin" "$@" --json | diff - "testdata/golden/$name.json" >/dev/null || { echo "golden mismatch: $name"; exit 1; }
}
golden ls ls
golden tags tags
golden search-markdown search markdown
golden search-shared search shared
golden cat-welcome cat welcome
golden backlinks-plain backlinks plain
golden orphans orphans
golden broken broken
for n in ambiguous-source code-fences alias-source heading-source embed-source path-source; do
  golden "links-$n" links "$n"
done
echo "14 goldens match"

echo "== rebuild is identical"
"$bin" ls --json > "$scratch/before.json"
rm -r "$XDG_CACHE_HOME/mdkb"
"$bin" reindex --full >/dev/null
"$bin" ls --json | cmp - "$scratch/before.json"

echo "== live checks"
"$bin" cat nope >/dev/null 2>&1 && exit 1 || test $? -eq 2            # unknown id exits 2
"$bin" mv plain 'has#heading' >/dev/null 2>&1 && exit 1 || test $? -eq 5  # delimiter id refused
"$bin" --root "$scratch/nowhere" ls >/dev/null 2>&1 && exit 1 || test $? -eq 3  # uninitialised exits 3
"$bin" mv plain renamed/plain --json >/dev/null
test "$("$bin" broken --json | grep -o '"total":[0-9]*')" = '"total":3'   # no new broken rows after mv
idx="$("$bin" info --json | grep -o '"index_path":"[^"]*"' | cut -d'"' -f4)"
head -c 4096 /dev/urandom > "$idx"
"$bin" ls >/dev/null 2>&1 && exit 1 || test $? -eq 4                   # corrupt cache exits 4
"$bin" reindex --full >/dev/null                                        # and recovers
"$bin" daily --date 2026-01-01 --append seed >/dev/null
for i in 01 02 03 04 05 06 07 08 09 10; do "$bin" daily --date 2026-01-01 --append "m$i" >/dev/null & done; wait
test "$(grep -c '^m' "$MDKB_ROOT/10-Daily/2026-01-01.md")" -eq 10        # concurrent appends all survive
printf '%s\n' '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"accept","version":"0"}}}' \
  '{"jsonrpc":"2.0","method":"notifications/initialized"}' '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' > "$scratch/mcp.in"
tools="$( (cat "$scratch/mcp.in"; sleep 2) | "$bin" mcp | grep -o '"name":"[a-z_]*"' | sort -u | grep -vc '"mdkb"')"
test "$tools" -eq 6                                                      # six read tools without --allow-write
tools="$( (cat "$scratch/mcp.in"; sleep 2) | "$bin" mcp --allow-write | grep -o '"name":"[a-z_]*"' | sort -u | grep -vc '"mdkb"')"
test "$tools" -eq 8                                                      # eight with it

echo "accept: ok"
