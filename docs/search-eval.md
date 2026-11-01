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

> **TODO: run on Mac Studio with nomic-embed-text**
>
> Semantic hit@3 / MRR: pending.
>
> Hybrid hit@3 / MRR: pending.
>
> Hybrid ranking acceptance remains unverified until this block is replaced
> with measured results. No real-model improvement is claimed by the fake tests.
