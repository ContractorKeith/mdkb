package index

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/ContractorKeith/mdkb/internal/kb"
)

// Sync atomically applies changed files and removes vanished notes.
func (i *Index) Sync(ctx context.Context, root string, full bool) error {
	notes, err := kb.Walk(root)
	if err != nil {
		return err
	}
	tx, err := i.db.BeginTx(ctx, nil)
	if err != nil {
		return Error(err)
	}
	defer tx.Rollback()
	if full {
		if _, err = tx.ExecContext(ctx, `DROP TRIGGER embedding_delete; DROP TRIGGER embedding_update; DROP TABLE embeddings; DROP TABLE embedding_state; DROP TRIGGER notes_ai; DROP TRIGGER notes_ad; DROP TRIGGER notes_au; DROP TABLE notes_fts; DROP TABLE links; DROP TABLE link_sources; DROP TABLE tags; DROP TABLE notes;`+schema); err != nil {
			return Error(err)
		}
	}
	rows, err := tx.QueryContext(ctx, "SELECT path,mtime,size FROM notes ORDER BY path")
	if err != nil {
		return Error(err)
	}
	type stamp struct{ mtime, size int64 }
	old := map[string]stamp{}
	for rows.Next() {
		var path string
		var s stamp
		if err := rows.Scan(&path, &s.mtime, &s.size); err != nil {
			rows.Close()
			return Error(err)
		}
		old[path] = s
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return Error(err)
	}
	for _, n := range notes {
		s, exists := old[n.Path]
		delete(old, n.Path)
		var scanned int
		if err := tx.QueryRowContext(ctx, "SELECT count(*) FROM link_sources WHERE note_id=?", n.ID).Scan(&scanned); err != nil {
			return Error(err)
		}
		if exists && s.mtime == n.MTime && s.size == n.Size && scanned != 0 {
			continue
		}
		if err := upsertNote(ctx, tx, n); err != nil {
			return Error(err)
		}
	}
	for path := range old {
		if _, err = tx.ExecContext(ctx, "DELETE FROM links WHERE from_id=(SELECT id FROM notes WHERE path=?)", path); err != nil {
			return Error(err)
		}
		if _, err = tx.ExecContext(ctx, "DELETE FROM link_sources WHERE note_id=(SELECT id FROM notes WHERE path=?)", path); err != nil {
			return Error(err)
		}
		if _, err = tx.ExecContext(ctx, "DELETE FROM tags WHERE note_id=(SELECT id FROM notes WHERE path=?)", path); err != nil {
			return Error(err)
		}
		if _, err = tx.ExecContext(ctx, "DELETE FROM notes WHERE path=?", path); err != nil {
			return Error(err)
		}
	}
	if err := resolveLinks(ctx, tx, notes); err != nil {
		return Error(err)
	}
	return Error(tx.Commit())
}

// SyncNote reindexes one written note and reconsiders graph resolution.
func (i *Index) SyncNote(ctx context.Context, root, id string) error {
	n, _, err := kb.Read(root, id)
	if err != nil {
		return err
	}
	tx, err := i.db.BeginTx(ctx, nil)
	if err != nil {
		return Error(err)
	}
	defer tx.Rollback()
	if err := upsertNote(ctx, tx, n); err != nil {
		return Error(err)
	}
	notes, err := resolutionNotesTx(ctx, tx)
	if err != nil {
		return Error(err)
	}
	if err := resolveLinks(ctx, tx, notes); err != nil {
		return Error(err)
	}
	return Error(tx.Commit())
}

func upsertNote(ctx context.Context, tx *sql.Tx, n kb.Note) error {
	if err := replaceLinks(ctx, tx, n); err != nil {
		return err
	}
	aliases, _ := json.Marshal(n.Aliases)
	tags, _ := json.Marshal(n.Tags)
	_, err := tx.ExecContext(ctx, `INSERT INTO notes(id,path,title,aliases,tags,created,updated,mtime,size,body,problem) VALUES(?,?,?,?,?,?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET path=excluded.path,title=excluded.title,aliases=excluded.aliases,tags=excluded.tags,created=excluded.created,updated=excluded.updated,mtime=excluded.mtime,size=excluded.size,body=excluded.body,problem=excluded.problem`, n.ID, n.Path, n.Title, string(aliases), string(tags), n.Created, n.Updated, n.MTime, n.Size, n.Body, n.Problem)
	if err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, "DELETE FROM tags WHERE note_id=?", n.ID); err != nil {
		return err
	}
	for _, tag := range n.Tags {
		if _, err = tx.ExecContext(ctx, "INSERT OR IGNORE INTO tags(note_id,name) VALUES(?,?)", n.ID, tag); err != nil {
			return err
		}
	}
	return nil
}
