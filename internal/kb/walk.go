package kb

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

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
		if ValidateID(strings.TrimSuffix(filepath.ToSlash(rel), ".md"), false) != nil {
			return nil
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
