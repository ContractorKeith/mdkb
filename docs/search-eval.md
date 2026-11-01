# Fixture search evaluation

The ten judgments in testdata/eval/queries.json come from reading the fixture:
shared basenames, Markdown examples, the orphan, frontmatter, anchors, captures,
ambiguous links, the Café, the isolated note, and recoverable malformed YAML.
Expected lists contain up to three relevant IDs; one-answer paraphrases have
one relevant ID. Hit@3 means any relevant ID occurs in the first three results.
MRR uses the first relevant rank within the top 50, or zero if absent.

Three queries deliberately have no FTS word overlap with their intended note:
“espresso shop” → spaces and unicode, “isolated document” → orphan, and
“damaged YAML recovery” → bad-frontmatter. These test retrieval by meaning;
the other seven check that keyword strengths survive fusion.

## Linux keyword baseline — 2026-11-01

Run against a freshly initialized copy of testdata/kb (22 notes), default FTS5:

| Query | Hit@3 | MRR |
|---|---:|---:|
| shared | 1 | 1 |
| markdown | 1 | 1 |
| orphan | 1 | 1 |
| frontmatter | 1 | 1 |
| anchor | 1 | 1 |
| capture | 1 | 1 |
| ambiguous | 1 | 1 |
| espresso shop | 0 | 0 |
| isolated document | 0 | 0 |
| damaged YAML recovery | 0 | 0 |
| **Mean** | **0.7000** | **0.7000** |

No Ollama is installed on the Linux implementation machine. HTTP tests use
httptest.Server with deterministic 64-dimensional vectors and controlled synonyms;
those verify ranking, HTTP batching and fallback, but are not real-model evidence.

## Run on Mac Studio

Install/run Ollama on that machine and pull the model there:

```sh
ollama pull nomic-embed-text
go build ./cmd/mdkb
# Use a new temporary directory, never an existing KB or vault.
eval_root="$(mktemp -d)"
eval_cache="$(mktemp -d)"
./mdkb init "$eval_root"
cp -R testdata/kb/. "$eval_root/"
export MDKB_ROOT="$eval_root" XDG_CACHE_HOME="$eval_cache"
export MDKB_SEMANTIC=1 MDKB_EMBED_MODEL=nomic-embed-text
./mdkb embed --json
./mdkb eval --queries testdata/eval/queries.json
./mdkb eval --json > /tmp/mdkb-search-eval.json
```

Record Ollama version, model digest (ollama show), machine, date, per-mode means,
and per-query results. Compare hybrid hit@3 and MRR with keyword 0.7000/0.7000;
investigate regressions before claiming M5's live ranking acceptance is met.
Use --force when repeating embedding with changed model weights under the same name.

## Mac Studio real-model results — 2026-11-01

Mac Studio (`Mac14,13`, macOS 26.6, arm64, Go 1.27.3), Ollama 0.40.13,
`nomic-embed-text` 137M F16, digest
`0a109f422b47e3a30ba2b10eca18548e944e8a23073ee3f3e947efcf3c45e59f`, mdkb at
`f357ec3` (v0.1.0). Fresh fixture copy: `embed` stored 22 notes / 23 chunks in 1.1 s.

| Mode | Hit@3 | MRR |
|---|---:|---:|
| keyword | 0.7000 | 0.7000 |
| semantic | 0.9000 | 0.9067 |
| hybrid | 0.9000 | 0.9067 |

Hybrid is above the keyword baseline on both metrics, so the M5 acceptance is met.
The two no-overlap paraphrases "espresso shop" and "damaged YAML recovery" now rank
their note first; "isolated document" still misses (the orphan sits at rank 15 for
the real model too). Repeat with `./mdkb eval --json` after any change to chunking,
the model or the fixture, and update this table.
