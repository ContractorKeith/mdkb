package kb

import (
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var writeMu sync.Mutex

func validID(id string) bool {
	if id == "" || filepath.IsAbs(id) || strings.ContainsAny(id, "\\\x00") || strings.Contains(id, "..") {
		return false
	}
	for _, part := range strings.Split(id, "/") {
		if part == "" || strings.HasPrefix(part, ".") {
			return false
		}
	}
	return !strings.HasSuffix(id, ".md")
}

type WriteResult struct {
	ID    string `json:"id" jsonschema:"Note ID relative to root without .md"`
	Path  string `json:"path" jsonschema:"Relative Markdown path"`
	Bytes int    `json:"bytes" jsonschema:"Total bytes in the resulting file"`
	Mode  string `json:"mode" jsonschema:"Applied write mode"`
}

// Write atomically publishes a note; cache updates belong to the application layer.
func Write(root, id, content, mode string) (WriteResult, error) {
	writeMu.Lock()
	defer writeMu.Unlock()
	result, err := write(root, id, content, mode)
	if err != nil {
		return WriteResult{}, refuse(err)
	}
	return result, nil
}

func write(root, id, content, mode string) (WriteResult, error) {
	if !validID(id) {
		return WriteResult{}, fmt.Errorf("invalid note id: %q", id)
	}
	if mode != "create" && mode != "append" && mode != "overwrite" {
		return WriteResult{}, fmt.Errorf("invalid write mode: %s", mode)
	}
	initialized, err := hasMarker(root)
	if err != nil {
		return WriteResult{}, err
	}
	if !initialized {
		return WriteResult{}, errors.New("KB not initialized")
	}
	dir, err := os.OpenRoot(root)
	if err != nil {
		return WriteResult{}, err
	}
	defer func() { dir.Close() }()
	parts := strings.Split(id, "/")
	// Pin each directory handle so replacement by a symlink cannot redirect a write.
	for _, part := range parts[:len(parts)-1] {
		stat, err := dir.Lstat(part)
		if errors.Is(err, os.ErrNotExist) {
			if err = dir.Mkdir(part, 0755); err != nil {
				return WriteResult{}, err
			}
			stat, err = dir.Lstat(part)
		}
		if err != nil {
			return WriteResult{}, err
		}
		if !stat.IsDir() || stat.Mode()&os.ModeSymlink != 0 {
			return WriteResult{}, errors.New("directory symlink or non-directory refused")
		}
		next, err := dir.OpenRoot(part)
		if err != nil {
			return WriteResult{}, err
		}
		dir.Close()
		dir = next
	}
	name := parts[len(parts)-1] + ".md"
	stat, err := dir.Lstat(name)
	exists := err == nil
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return WriteResult{}, err
	}
	if exists && !stat.Mode().IsRegular() {
		return WriteResult{}, errors.New("non-regular note refused")
	}
	if mode == "create" && exists {
		return WriteResult{}, os.ErrExist
	}
	if mode != "create" && !exists {
		return WriteResult{}, os.ErrNotExist
	}
	if mode == "append" {
		raw, err := dir.ReadFile(name)
		if err != nil {
			return WriteResult{}, err
		}
		separator := "\n\n"
		if strings.HasSuffix(string(raw), "\n") {
			separator = "\n"
		}
		content = string(raw) + separator + content
	}
	temp := fmt.Sprintf(".mdkb-%x.tmp", randomBytes())
	f, err := dir.OpenFile(temp, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return WriteResult{}, err
	}
	defer dir.Remove(temp)
	_, err = f.WriteString(content)
	if err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err != nil {
		return WriteResult{}, err
	}
	if closeErr != nil {
		return WriteResult{}, closeErr
	}
	if mode == "create" {
		// Link publishes without replacing a note concurrently created by another process.
		err = dir.Link(temp, name)
	} else {
		err = dir.Rename(temp, name)
	}
	if err != nil {
		return WriteResult{}, err
	}
	return WriteResult{ID: id, Path: id + ".md", Bytes: len(content), Mode: mode}, nil
}

func randomBytes() []byte {
	b := make([]byte, 16)
	// rand.Read is guaranteed to succeed.
	_, _ = rand.Read(b)
	return b
}
