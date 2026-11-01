package kb

import (
	"errors"
	"os"
	"path"
	"strings"

	"golang.org/x/text/unicode/norm"
)

// CanonicalID is the spelling used for comparisons and public note IDs.
func CanonicalID(id string) string { return norm.NFC.String(id) }

// resolveNotePath retains the on-disk spelling on normalization-sensitive filesystems.
func resolveNotePath(root *os.Root, rel string) (string, error) {
	parts := strings.Split(rel, "/")
	for i, part := range parts {
		candidate := strings.Join(parts[:i+1], "/")
		if _, err := root.Lstat(candidate); err == nil {
			continue
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
		parent, err := root.Open(path.Dir(candidate))
		if errors.Is(err, os.ErrNotExist) {
			return strings.Join(parts, "/"), nil
		}
		if err != nil {
			return "", err
		}
		names, err := parent.Readdirnames(-1)
		parent.Close()
		if err != nil {
			return "", err
		}
		for _, name := range names {
			if CanonicalID(name) == part {
				parts[i] = name
				break
			}
		}
	}
	return strings.Join(parts, "/"), nil
}

func notePath(root, id string) (string, error) {
	dir, err := os.OpenRoot(root)
	if err != nil {
		return "", err
	}
	defer dir.Close()
	return resolveNotePath(dir, id+".md")
}
