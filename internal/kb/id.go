package kb

import (
	"errors"
	"path/filepath"
	"strings"
)

// ValidateID shares path rules for reads and writes; writes refuse an accidental extension.
func ValidateID(id string, write bool) error {
	invalid := id == "" || filepath.IsAbs(id) || strings.ContainsAny(id, "\\\x00")
	for _, part := range strings.Split(id, "/") {
		invalid = invalid || part == "" || strings.HasPrefix(part, ".")
	}
	if write {
		invalid = invalid || strings.HasSuffix(id, ".md")
	}
	if invalid {
		return refuse(errors.New("invalid note id: " + id))
	}
	return nil
}
