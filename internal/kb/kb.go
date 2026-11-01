// Package kb owns all access to knowledge-base files.
package kb

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var ErrWriteRefused = errors.New("write refused")

// Details is the stable info response, including for an uninitialized directory.
type Details struct {
	Links        int    `json:"links" jsonschema:"links"`
	BrokenLinks  int    `json:"broken_links" jsonschema:"broken links"`
	Orphans      int    `json:"orphans" jsonschema:"orphans"`
	Problems     int    `json:"problems" jsonschema:"problems"`
	IndexedNotes int    `json:"indexed_notes" jsonschema:"indexed notes"`
	IndexPath    string `json:"index_path" jsonschema:"index path"`
	Root         string `json:"root" jsonschema:"root"`
	NoteCount    int    `json:"note_count" jsonschema:"note count"`
	Initialized  bool   `json:"initialized" jsonschema:"initialized"`
}

// Init creates a KB, leaving an existing KB untouched.
func Init(path string) error { return initialize(path, false) }

// PreviewInit performs the same safety checks as Init without writing.
func PreviewInit(path string) error { return initialize(path, true) }

func initialize(path string, dryRun bool) error {
	root, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	initialized, err := hasMarker(root)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrWriteRefused, err)
	}
	if initialized {
		return nil
	}
	entries, err := os.ReadDir(root)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("%w: %v", ErrWriteRefused, err)
	}
	if len(entries) > 0 {
		return fmt.Errorf("%w: %s is not empty and is not an mdkb KB", ErrWriteRefused, root)
	}
	if dryRun {
		return nil
	}
	if err := os.MkdirAll(root, 0755); err != nil {
		return refuse(err)
	}
	// Reuse an enclosing repository rather than creating a nested one.
	probe := exec.Command("git", "-C", root, "rev-parse", "--show-toplevel")
	if err := probe.Run(); err != nil {
		if output, err := exec.Command("git", "-C", root, "init", "--quiet").CombinedOutput(); err != nil {
			return refuse(fmt.Errorf("git init: %w: %s", err, strings.TrimSpace(string(output))))
		}
	}
	for _, dir := range []string{"00-Inbox", ".mdkb"} {
		if err := os.MkdirAll(filepath.Join(root, dir), 0755); err != nil {
			return refuse(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, ".gitignore"), []byte(".mdkb/\n.DS_Store\n"), 0644); err != nil {
		return refuse(err)
	}
	return refuse(os.WriteFile(filepath.Join(root, ".mdkb", "config.toml"), []byte("version = 1\n"), 0644))
}

func refuse(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%w: %v", ErrWriteRefused, err)
}

func hasMarker(root string) (bool, error) {
	marker, err := os.Stat(filepath.Join(root, ".mdkb", "config.toml"))
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return marker.Mode().IsRegular(), nil
}

// Info counts Markdown files recursively, skipping dot-directories and child symlinks.
// A missing root is an error; an existing uninitialized root is inspectable.
func Info(root string) (Details, error) {
	absolute, err := filepath.Abs(root)
	if err != nil {
		return Details{}, err
	}
	result := Details{Root: absolute}
	stat, err := os.Stat(absolute)
	if err != nil {
		return result, err
	}
	if !stat.IsDir() {
		return result, fmt.Errorf("%s is not a directory", absolute)
	}
	result.Initialized, err = hasMarker(absolute)
	if err != nil {
		return result, err
	}
	// WalkDir does not follow even a root symlink; resolve it explicitly.
	walkRoot, err := filepath.EvalSymlinks(absolute)
	if err != nil {
		return result, err
	}
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
		if entry.Type().IsRegular() && filepath.Ext(entry.Name()) == ".md" {
			result.NoteCount++
			raw, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			note, _ := Parse(entry.Name(), raw)
			if note.Problem != "" {
				result.Problems++
			}
		}
		return nil
	})
	return result, err
}
