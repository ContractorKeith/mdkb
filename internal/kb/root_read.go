package kb

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"
	"syscall"
)

var errUnsafeNote = errors.New("unsafe note path")

// readNoteFile refuses child symlinks and pins the opened file for data and metadata.
func readNoteFile(root *os.Root, rel string) ([]byte, fs.FileInfo, error) {
	parts := strings.Split(rel, "/")
	for i := range parts {
		stat, err := root.Lstat(strings.Join(parts[:i+1], "/"))
		if err != nil {
			return nil, nil, err
		}
		if stat.Mode()&os.ModeSymlink != 0 || (i < len(parts)-1 && !stat.IsDir()) {
			return nil, nil, fmt.Errorf("%w: child symlink or non-directory refused: %s", errUnsafeNote, rel)
		}
	}
	// O_NOFOLLOW closes the final-component swap window; Root confines every parent.
	f, err := root.OpenFile(rel, os.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil {
		return nil, nil, err
	}
	if !stat.Mode().IsRegular() {
		return nil, nil, fmt.Errorf("%w: non-regular note refused: %s", errUnsafeNote, rel)
	}
	raw, err := io.ReadAll(f)
	return raw, stat, err
}
