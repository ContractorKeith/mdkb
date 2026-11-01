package kb

import (
	"errors"
	"os"
	"strings"
)

// PreviewWrite validates a proposed write without creating directories or files.
func PreviewWrite(root, id, mode string) error {
	if ValidateID(id, true) != nil {
		return refuse(errors.New("invalid note id: " + id))
	}
	ok, err := hasMarker(root)
	if err != nil {
		return refuse(err)
	}
	if !ok {
		return refuse(errors.New("KB not initialized"))
	}
	dir, err := os.OpenRoot(root)
	if err != nil {
		return refuse(err)
	}
	defer dir.Close()
	parts := strings.Split(id+".md", "/")
	for i := range parts {
		stat, err := dir.Lstat(strings.Join(parts[:i+1], "/"))
		if errors.Is(err, os.ErrNotExist) {
			if mode == "create" {
				return nil
			}
			return refuse(err)
		}
		if err != nil {
			return refuse(err)
		}
		if stat.Mode()&os.ModeSymlink != 0 {
			return refuse(errors.New("symlink refused"))
		}
		if i < len(parts)-1 && !stat.IsDir() {
			return refuse(errors.New("non-directory refused"))
		}
		if i == len(parts)-1 {
			if !stat.Mode().IsRegular() {
				return refuse(errors.New("non-regular note refused"))
			}
			if mode == "create" {
				return refuse(os.ErrExist)
			}
		}
	}
	return nil
}
