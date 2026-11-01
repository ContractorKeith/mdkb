package kb

import (
	"os"
	"strings"
)

// noteDir pins each parent directory and refuses child symlinks.
func noteDir(root, id string, create bool) (*os.Root, string, error) {
	dir, err := os.OpenRoot(root)
	if err != nil {
		return nil, "", err
	}
	rel, err := resolveNotePath(dir, id+".md")
	if err != nil {
		dir.Close()
		return nil, "", err
	}
	parts := strings.Split(strings.TrimSuffix(rel, ".md"), "/")
	for _, part := range parts[:len(parts)-1] {
		stat, err := dir.Lstat(part)
		if os.IsNotExist(err) && create {
			err = dir.Mkdir(part, 0755)
			if err == nil {
				stat, err = dir.Lstat(part)
			}
		}
		if err != nil {
			dir.Close()
			return nil, "", err
		}
		if !stat.IsDir() || stat.Mode()&os.ModeSymlink != 0 {
			dir.Close()
			return nil, "", os.ErrPermission
		}
		next, err := dir.OpenRoot(part)
		dir.Close()
		if err != nil {
			return nil, "", err
		}
		dir = next
	}
	return dir, parts[len(parts)-1] + ".md", nil
}

func renameNote(root, old, new string) error {
	source, _, err := noteDir(root, old, false)
	if err != nil {
		return err
	}
	defer source.Close()
	dest, newName, err := noteDir(root, new, true)
	if err != nil {
		return err
	}
	defer dest.Close()
	// Reserve the destination exclusively before rename, refusing concurrent creates.
	f, err := dest.OpenFile(newName, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	f.Close()
	dir, err := os.OpenRoot(root)
	if err != nil {
		dest.Remove(newName)
		return err
	}
	defer dir.Close()
	oldPath, err := resolveNotePath(dir, old+".md")
	if err != nil {
		dest.Remove(newName)
		return err
	}
	newPath, err := resolveNotePath(dir, new+".md")
	if err != nil {
		dest.Remove(newName)
		return err
	}
	if err := dir.Rename(oldPath, newPath); err != nil {
		dest.Remove(newName)
		return err
	}
	return nil
}
