# Golden fixture KB

Twenty notes plus this README (21 Markdown files). This is source data, not an
initialized KB: initialize a temporary directory and copy these fixtures into it.
Do not normalize the deliberate ambiguities or broken links.

| File | Exercises |
| --- | --- |
| 00-Inbox/capture.md | Tags only; filename fallback; inbox folder |
| welcome.md | Title, tags, aliases, created and updated; basic links |
| plain.md | No frontmatter or H1; heading link |
| h1-only.md | H1 title fallback; display alias |
| filename-title.md | Filename title fallback |
| empty-frontmatter.md | Empty metadata block; H1 fallback |
| multiline-tags.md | Block lists, spaced tag, frontmatter alias target |
| alias-source.md | Display aliases and frontmatter alias resolution |
| headings.md | H1 and H2 anchor targets |
| heading-source.md | Heading anchors, including anchor plus display alias |
| embed-source.md | Note and heading embeds |
| code-fences.md | Backtick and tilde fences containing ignored links; real link outside |
| broken-source.md | Missing basename and missing explicit path |
| ambiguous-source.md | Ambiguous basename and exact paths to both candidates |
| projects/shared.md | First shared basename; alias; path link |
| archive/shared.md | Second shared basename; empty tags |
| nested/deep/path-target.md | Deep nesting and unique alias resolution |
| path-source.md | Exact path and unique basename resolution |
| spaces and unicode.md | Spaces in filename, Unicode, quoted colon in title |
| orphan.md | No incoming or outgoing links |
| README.md | Fixture manifest; another unlinked Markdown file |

There are intentionally no links to the manifest or the orphan. Other notes may
also lack incoming links; future golden reports should reflect the whole graph.
