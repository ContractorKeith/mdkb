package kb

import "strings"

// LinkRef retains the spelling inside brackets and a body-relative, one-based line.
type LinkRef struct {
	Raw, Target, Heading, Alias string
	Start, End                  int
	Embed                       bool
	Line                        int
}

// ExtractLinks scans wikilinks outside fences and matched inline code spans.
func ExtractLinks(body string) []LinkRef {
	result := []LinkRef{}
	fence := byte(0)
	inline := 0
	offset := 0
	for lineNo, line := range strings.Split(body, "\n") {
		lineOffset := offset
		offset += len(line) + 1
		trimmed := strings.TrimSpace(line)
		if inline == 0 && fenceLine(trimmed) {
			if fence == 0 {
				fence = trimmed[0]
			} else if fence == trimmed[0] {
				fence = 0
			}
			continue
		}
		if fence != 0 {
			continue
		}
		for p := 0; p < len(line); {
			if line[p] == '`' {
				end := p + 1
				for end < len(line) && line[end] == '`' {
					end++
				}
				width := end - p
				if inline == width {
					inline = 0
				} else if inline == 0 && hasCodeEnd(body[lineOffset+end:], width) {
					inline = width
				}
				p = end
				continue
			}
			if inline != 0 || !strings.HasPrefix(line[p:], "[[") {
				p++
				continue
			}
			end := strings.Index(line[p+2:], "]]")
			if end < 0 {
				break
			}
			end += p + 2
			raw := line[p+2 : end]
			ref := LinkRef{Raw: raw, Start: lineOffset + p + 2, End: lineOffset + end, Embed: p > 0 && line[p-1] == '!', Line: lineNo + 1}
			target, alias, _ := strings.Cut(raw, "|")
			target, heading, _ := strings.Cut(target, "#")
			ref.Target, ref.Heading, ref.Alias = strings.TrimSpace(target), strings.TrimSpace(heading), strings.TrimSpace(alias)
			if ref.Target != "" {
				result = append(result, ref)
			}
			p = end + 2
		}
	}
	return result
}

func hasCodeEnd(s string, width int) bool {
	for p := 0; p < len(s); {
		if s[p] != '`' {
			p++
			continue
		}
		end := p + 1
		for end < len(s) && s[end] == '`' {
			end++
		}
		if end-p == width {
			return true
		}
		p = end
	}
	return false
}
