package kb

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// Read resolves one ID without traversing unrelated notes. Child symlinks are refused.
func Read(root, id string) (Note, string, error) {
	refused := func() (Note, string, error) { return Note{}, "", NoteNotFound(id) }
	failed := func(err error) (Note, string, error) {
		if errors.Is(err, os.ErrNotExist) {
			return refused()
		}
		return Note{}, "", err
	}
	if !validID(id) {
		return refused()
	}
	clean := filepath.Clean(filepath.FromSlash(id))
	absolute, err := filepath.Abs(root)
	if err != nil {
		return failed(err)
	}
	path := filepath.Join(absolute, clean+".md")
	rel, err := filepath.Rel(absolute, path)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return refused()
	}
	dir, err := os.OpenRoot(absolute)
	if err != nil {
		return failed(err)
	}
	defer dir.Close()
	parts := strings.Split(rel, string(filepath.Separator))
	for i, part := range parts {
		if i < len(parts)-1 && strings.HasPrefix(part, ".") {
			return refused()
		}
		stat, err := dir.Lstat(filepath.Join(parts[:i+1]...))
		if err != nil {
			return failed(err)
		}
		if stat.Mode()&os.ModeSymlink != 0 {
			return refused()
		}
		if i < len(parts)-1 {
			if !stat.IsDir() {
				return refused()
			}
		} else if !stat.Mode().IsRegular() {
			return refused()
		}
	}
	// Root.ReadFile also enforces containment if a component changes after validation.
	raw, err := dir.ReadFile(rel)
	if err != nil {
		return failed(err)
	}
	n, err := Parse(rel, raw)
	if err != nil {
		return failed(err)
	}
	stat, err := dir.Stat(rel)
	if err != nil {
		return failed(err)
	}
	n.MTime, n.Size = stat.ModTime().UnixNano(), stat.Size()
	return n, string(raw), nil
}
