package kb

import (
	"fmt"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Note is a Markdown file; Path and ID are relative to the KB root.
type Note struct {
	ID      string   `json:"id"`
	Path    string   `json:"path"`
	Title   string   `json:"title"`
	Aliases []string `json:"aliases"`
	Tags    []string `json:"tags"`
	Created string   `json:"created"`
	Updated string   `json:"updated"`
	MTime   int64    `json:"mtime"`
	Size    int64    `json:"size"`
	Problem string   `json:"problem,omitempty"`
	Body    string   `json:"body"`
}

type stringList []string

func (s *stringList) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind == yaml.ScalarNode {
		var value string
		if err := n.Decode(&value); err != nil {
			return err
		}
		*s = []string{value}
		return nil
	}
	var values []string
	if err := n.Decode(&values); err != nil {
		return err
	}
	*s = values
	return nil
}

// Parse removes optional frontmatter, retaining the Markdown body verbatim.
func Parse(path string, raw []byte) (Note, error) {
	n := Note{ID: strings.TrimSuffix(filepath.ToSlash(path), ".md"), Path: filepath.ToSlash(path), Tags: []string{}, Aliases: []string{}, Body: string(raw)}
	lines := strings.SplitAfter(string(raw), "\n")
	if len(lines) > 0 && strings.TrimSpace(lines[0]) == "---" {
		end := -1
		for i := 1; i < len(lines); i++ {
			if strings.TrimSpace(lines[i]) == "---" {
				end = i
				break
			}
		}
		if end < 0 {
			n.Problem = "unterminated frontmatter"
		} else if err := parseMetadata(&n, strings.Join(lines[1:end], "")); err != nil {
			n.Problem = fmt.Sprintf("invalid frontmatter: %v", err)
		} else {
			n.Body = strings.Join(lines[end+1:], "")
		}
	}
	if n.Title == "" {
		fence := byte(0)
		for _, line := range strings.Split(n.Body, "\n") {
			line = strings.TrimSpace(line)
			if fenceLine(line) {
				if fence == 0 {
					fence = line[0]
				} else if fence == line[0] {
					fence = 0
				}
				continue
			}
			if fence == 0 && strings.HasPrefix(line, "# ") {
				n.Title = strings.TrimSpace(line[2:])
				trimmed := strings.TrimRight(n.Title, "#")
				if trimmed != n.Title && strings.HasSuffix(trimmed, " ") {
					n.Title = strings.TrimSpace(trimmed)
				}
				break
			}
		}
	}
	if n.Title == "" {
		n.Title = strings.TrimSuffix(filepath.Base(path), ".md")
	}
	return n, nil
}

// parseMetadata applies metadata only after the complete block decodes.
func parseMetadata(n *Note, block string) error {
	var meta struct {
		Title   string
		Tags    stringList
		Aliases stringList
		Created string
		Updated string
	}
	if err := yaml.Unmarshal([]byte(block), &meta); err != nil {
		return err
	}
	n.Title, n.Created, n.Updated = meta.Title, meta.Created, meta.Updated
	if meta.Tags != nil {
		n.Tags = meta.Tags
	}
	if meta.Aliases != nil {
		n.Aliases = meta.Aliases
	}
	return nil
}
