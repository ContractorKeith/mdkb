package kb

import (
	"fmt"
	"strings"
)

// RewriteLinks preserves resolved destinations across a move, retaining link suffixes.
func RewriteLinks(body, old, new string, before, after []Note) (string, int) {
	return rewriteLinks(body, CanonicalID(old), CanonicalID(new), NewResolver(before), NewResolver(after))
}

func rewriteLinks(body, old, new string, before, after *Resolver) (string, int) {
	refs := ExtractLinks(body)
	count := 0
	for i := len(refs) - 1; i >= 0; i-- {
		ref := refs[i]
		resolved, reason := before.ResolveLink(ref.Target)
		if reason != "" {
			continue
		}
		next, nextReason := after.ResolveLink(ref.Target)
		replacement := resolved
		if resolved == old {
			replacement = new
			exact := strings.TrimSuffix(ref.Target, ".md") == old || ref.Target == old
			if !exact && nextReason == "" && next == new {
				continue
			}
		} else if nextReason == "" && next == resolved {
			continue
		}
		end := strings.IndexAny(ref.Raw, "#|")
		if end < 0 {
			end = len(ref.Raw)
		}
		body = body[:ref.Start] + replacement + ref.Raw[end:] + body[ref.End:]
		count++
	}
	return body, count
}

// UnifiedDiff emits a valid whole-file unified hunk without external tools.
func UnifiedDiff(oldPath, newPath, before, after string) string {
	lines := func(s string) []string {
		if s == "" {
			return nil
		}
		return strings.SplitAfter(s, "\n")
	}
	a, b := lines(before), lines(after)
	if len(a) > 0 && a[len(a)-1] == "" {
		a = a[:len(a)-1]
	}
	if len(b) > 0 && b[len(b)-1] == "" {
		b = b[:len(b)-1]
	}
	var out strings.Builder
	fmtDiffHeader(&out, oldPath, newPath, len(a), len(b))
	for _, group := range []struct {
		prefix string
		lines  []string
	}{{"-", a}, {"+", b}} {
		for _, line := range group.lines {
			out.WriteString(group.prefix + line)
			if !strings.HasSuffix(line, "\n") {
				out.WriteString("\n\\ No newline at end of file\n")
			}
		}
	}
	return out.String()
}

func fmtDiffHeader(out *strings.Builder, old, new string, a, b int) {
	aStart, bStart := 1, 1
	if a == 0 {
		aStart = 0
	}
	if b == 0 {
		bStart = 0
	}
	fmt.Fprintf(out, "--- a/%s\n+++ b/%s\n@@ -%d,%d +%d,%d @@\n", old, new, aStart, a, bStart, b)
}
