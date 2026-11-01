// Package index owns the rebuildable SQLite cache and all SQL access.
package index

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

var ErrIndex = errors.New("index error")

type Index struct {
	db   *sql.DB
	Path string
}

func Error(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%w: %v", ErrIndex, err)
}

func Open(root string) (*Index, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, Error(err)
	}
	cache := os.Getenv("XDG_CACHE_HOME")
	if cache == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, Error(err)
		}
		cache = filepath.Join(home, ".cache")
	}
	hash := sha256.Sum256([]byte(root))
	path := filepath.Join(cache, "mdkb", fmt.Sprintf("%x.db", hash[:8]))
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, Error(err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, Error(err)
	}
	db.SetMaxOpenConns(1)
	idx := &Index{db: db, Path: path}
	if _, err := db.Exec("PRAGMA busy_timeout=5000;" + schema); err != nil {
		db.Close()
		return nil, Error(err)
	}
	return idx, nil
}
func (i *Index) Close() error { return i.db.Close() }
func (i *Index) Count(ctx context.Context) (int, error) {
	var n int
	err := i.db.QueryRowContext(ctx, "SELECT count(*) FROM notes").Scan(&n)
	return n, Error(err)
}

const schema = `
CREATE TABLE IF NOT EXISTS notes (id TEXT PRIMARY KEY, path TEXT NOT NULL UNIQUE, title TEXT NOT NULL, aliases TEXT NOT NULL, tags TEXT NOT NULL, created TEXT NOT NULL, updated TEXT NOT NULL, mtime INTEGER NOT NULL, size INTEGER NOT NULL, body TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS tags (note_id TEXT NOT NULL, name TEXT NOT NULL, PRIMARY KEY(note_id,name));
CREATE VIRTUAL TABLE IF NOT EXISTS notes_fts USING fts5(title, body, content='notes', content_rowid='rowid');
CREATE TRIGGER IF NOT EXISTS notes_ai AFTER INSERT ON notes BEGIN INSERT INTO notes_fts(rowid,title,body) VALUES(new.rowid,new.title,new.body); END;
CREATE TRIGGER IF NOT EXISTS notes_ad AFTER DELETE ON notes BEGIN INSERT INTO notes_fts(notes_fts,rowid,title,body) VALUES('delete',old.rowid,old.title,old.body); END;
CREATE TRIGGER IF NOT EXISTS notes_au AFTER UPDATE ON notes BEGIN INSERT INTO notes_fts(notes_fts,rowid,title,body) VALUES('delete',old.rowid,old.title,old.body); INSERT INTO notes_fts(rowid,title,body) VALUES(new.rowid,new.title,new.body); END;
`
