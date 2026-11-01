package kb

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

// Read resolves one ID without traversing unrelated notes. Child symlinks are refused.
func Read(root, id string) (Note, string, error) {
	refused := func() (Note, string, error) { return Note{}, "", NoteNotFound(id) }
	failed := func(err error) (Note, string, error) {
		if errors.Is(err, os.ErrNotExist) || errors.Is(err, errUnsafeNote) || errors.Is(err, syscall.ELOOP) {
			return refused()
		}
		return Note{}, "", err
	}
	if ValidateID(id, false) != nil {
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
	raw, stat, err := readNoteFile(dir, filepath.ToSlash(rel))
	if err != nil {
		return failed(err)
	}
	n, err := Parse(rel, raw)
	if err != nil {
		return failed(err)
	}
	n.MTime, n.Size = stat.ModTime().UnixNano(), stat.Size()
	return n, string(raw), nil
}
