package index

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// CachePath hashes the physical root so filesystem aliases share one cache.
func CachePath(root string) (string, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return "", Error(err)
	}
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		return "", Error(err)
	}
	cache := os.Getenv("XDG_CACHE_HOME")
	if cache == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", Error(err)
		}
		cache = filepath.Join(home, ".cache")
	}
	hash := sha256.Sum256([]byte(root))
	return filepath.Join(cache, "mdkb", fmt.Sprintf("%x.db", hash[:8])), nil
}

type unreadableError struct {
	path  string
	cause error
}

func (e *unreadableError) Error() string {
	return fmt.Sprintf("index error: index at %s is unreadable (%v); run mdkb reindex --full", e.path, e.cause)
}
func (e *unreadableError) Unwrap() error { return ErrIndex }

// IsUnreadable distinguishes open/schema failures from ordinary query errors.
func IsUnreadable(err error) bool {
	var unreadable *unreadableError
	return errors.As(err, &unreadable)
}

// OpenForReindex discards an unreadable cache and all SQLite sidecars before retrying.
func OpenForReindex(root string) (*Index, error) {
	idx, err := Open(root)
	var unreadable *unreadableError
	if !errors.As(err, &unreadable) {
		return idx, err
	}
	for _, suffix := range []string{"", "-wal", "-shm"} {
		if err := os.Remove(unreadable.path + suffix); err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, Error(err)
		}
	}
	return Open(root)
}
