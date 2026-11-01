package embed

import "strings"

// Chunks prefers paragraph ends near 200 words and carries forty words forward.
// Oversized paragraphs split at the word limit. Even an empty note embeds its title.
func Chunks(title, body string) []string {
	words := []string{}
	ends := map[int]bool{}
	for _, p := range strings.Split(strings.ReplaceAll(body, "\r\n", "\n"), "\n\n") {
		words = append(words, strings.Fields(p)...)
		ends[len(words)] = true
	}
	chunks := []string{}
	for start := 0; start < len(words); {
		end := min(start+200, len(words))
		if end < len(words) {
			for n := end; n >= start+160; n-- {
				if ends[n] {
					end = n
					break
				}
			}
		}
		chunks = append(chunks, strings.Join(words[start:end], " "))
		if end == len(words) {
			break
		}
		start = end - 40
	}
	if len(chunks) == 0 {
		chunks = append(chunks, "")
	}
	chunks[0] = title + "\n" + chunks[0]
	return chunks
}
