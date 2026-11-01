package kb

import "os"

// NoteNotFound gives file and index reads the same public error and exit identity.
func NoteNotFound(id string) error { return noteNotFound(id) }

type noteNotFound string

func (e noteNotFound) Error() string { return "note not found: " + string(e) }
func (e noteNotFound) Unwrap() error { return os.ErrNotExist }
