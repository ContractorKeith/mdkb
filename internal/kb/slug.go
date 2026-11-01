package kb

import (
	"strings"
	"time"
	"unicode"
)

func Slug(text string) string {
	var result []rune
	for _, r := range strings.ToLower(text) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			result = append(result, r)
		} else if len(result) > 0 && result[len(result)-1] != '-' {
			result = append(result, '-')
		}
		if len(result) == 40 {
			break
		}
	}
	value := strings.Trim(string(result), "-")
	if value == "" {
		return "note"
	}
	return value
}

// CaptureContent shares timestamp, filename and frontmatter rules across front ends.
func CaptureContent(text, title string, now time.Time) (string, string, error) {
	source := title
	if source == "" {
		source = text
	}
	id := "00-Inbox/" + now.Format("2006-01-02-150405") + "-" + Slug(source)
	content, err := Content(text, &Frontmatter{Title: title, Created: now.Format(time.RFC3339)})
	return id, content, err
}
