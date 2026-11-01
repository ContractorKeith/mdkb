package kb

import (
	"fmt"
	"io/fs"
	"os"
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
			return n, fmt.Errorf("%s: unterminated frontmatter", path)
		}
		var meta struct {
			Title   string
			Tags    stringList
			Aliases stringList
			Created string
			Updated string
		}
		if err := yaml.Unmarshal([]byte(strings.Join(lines[1:end], "")), &meta); err != nil {
			return n, fmt.Errorf("%s: %w", path, err)
		}
		n.Title, n.Created, n.Updated = meta.Title, meta.Created, meta.Updated
		if meta.Tags != nil {
			n.Tags = meta.Tags
		}
		if meta.Aliases != nil {
			n.Aliases = meta.Aliases
		}
		n.Body = strings.Join(lines[end+1:], "")
	}
	if n.Title == "" {
		fence := byte(0)
		for _, line := range strings.Split(n.Body, "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "```") || strings.HasPrefix(line, "~~~") {
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

// Walk skips dot-directories and child symlinks, matching Info's file selection.
func Walk(root string) ([]Note, error) {
	walkRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return nil, err
	}
	notes := []Note{}
	err = filepath.WalkDir(walkRoot, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if path != walkRoot && strings.HasPrefix(entry.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if !entry.Type().IsRegular() || filepath.Ext(path) != ".md" {
			return nil
		}
		stat, err := entry.Info()
		if err != nil {
			return err
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(walkRoot, path)
		if err != nil {
			return err
		}
		note, err := Parse(rel, raw)
		if err != nil {
			return err
		}
		note.MTime, note.Size = stat.ModTime().UnixNano(), stat.Size()
		notes = append(notes, note)
		return nil
	})
	return notes, err
}

// Read returns the raw file for an ID selected from the safe walk.
func Read(root, id string) (Note, string, error) {
	notes, err := Walk(root)
	if err != nil {
		return Note{}, "", err
	}
	for _, n := range notes {
		if n.ID == id {
			raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(n.Path)))
			return n, string(raw), err
		}
	}
	return Note{}, "", fmt.Errorf("%s: %w", id, os.ErrNotExist)
}
