package kb

import (
	"io/fs"
	"os"
	"path"
	"strings"
)

// Walk skips dot-directories, invalid IDs and child symlinks, matching Info.
func Walk(root string) ([]Note, error) {
	notes, _, err := walkNotes(root)
	return notes, err
}

func walkNotes(root string) ([]Note, int, error) {
	dir, err := os.OpenRoot(root)
	if err != nil {
		return nil, 0, err
	}
	defer dir.Close()
	notes := []Note{}
	problems := 0
	err = fs.WalkDir(dir.FS(), ".", func(rel string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if rel != "." && strings.HasPrefix(entry.Name(), ".") {
				return fs.SkipDir
			}
			return nil
		}
		if !entry.Type().IsRegular() || path.Ext(rel) != ".md" {
			return nil
		}
		if ValidateID(strings.TrimSuffix(rel, ".md"), false) != nil {
			problems++
			return nil
		}
		raw, stat, err := readNoteFile(dir, rel)
		if err != nil {
			return err
		}
		note, err := Parse(rel, raw)
		if err != nil {
			return err
		}
		note.MTime, note.Size = stat.ModTime().UnixNano(), stat.Size()
		notes = append(notes, note)
		if note.Problem != "" {
			problems++
		}
		return nil
	})
	return notes, problems, err
}
