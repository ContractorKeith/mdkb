package index

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/ContractorKeith/mdkb/internal/kb"
)

type Link struct {
	Target   string  `json:"target" jsonschema:"target"`
	TargetID *string `json:"target_id" jsonschema:"target id"`
	Heading  string  `json:"heading" jsonschema:"heading"`
	Alias    string  `json:"alias" jsonschema:"alias"`
	Embed    bool    `json:"embed" jsonschema:"embed"`
	Line     int     `json:"line" jsonschema:"line"`
	Resolved bool    `json:"resolved" jsonschema:"resolved"`
}
type Outgoing struct {
	ID    string `json:"id" jsonschema:"id"`
	Links []Link `json:"links" jsonschema:"links"`
}
type Backlink struct {
	From  string `json:"from" jsonschema:"from"`
	Line  int    `json:"line" jsonschema:"line"`
	Alias string `json:"alias" jsonschema:"alias"`
	Embed bool   `json:"embed" jsonschema:"embed"`
}
type Incoming struct {
	ID        string     `json:"id" jsonschema:"id"`
	Backlinks []Backlink `json:"backlinks" jsonschema:"backlinks"`
}

func (i *Index) requireID(ctx context.Context, id string) error {
	var count int
	if err := i.db.QueryRowContext(ctx, "SELECT count(*) FROM notes WHERE id=?", id).Scan(&count); err != nil {
		return Error(err)
	}
	if count == 0 {
		return kb.NoteNotFound(id)
	}
	return nil
}
func (i *Index) Links(ctx context.Context, id string) (Outgoing, error) {
	id = kb.CanonicalID(id)
	result := Outgoing{ID: id, Links: []Link{}}
	if err := i.requireID(ctx, id); err != nil {
		return result, err
	}
	rows, err := i.db.QueryContext(ctx, `SELECT target_raw,target_id,heading,alias,kind='embed',line,target_id IS NOT NULL FROM links WHERE from_id=? ORDER BY line,rowid`, id)
	if err != nil {
		return result, Error(err)
	}
	defer rows.Close()
	for rows.Next() {
		var l Link
		if err := rows.Scan(&l.Target, &l.TargetID, &l.Heading, &l.Alias, &l.Embed, &l.Line, &l.Resolved); err != nil {
			return result, Error(err)
		}
		result.Links = append(result.Links, l)
	}
	return result, Error(rows.Err())
}
func (i *Index) Backlinks(ctx context.Context, id string) (Incoming, error) {
	id = kb.CanonicalID(id)
	result := Incoming{ID: id, Backlinks: []Backlink{}}
	if err := i.requireID(ctx, id); err != nil {
		return result, err
	}
	rows, err := i.db.QueryContext(ctx, `SELECT from_id,line,alias,kind='embed' FROM links WHERE target_id=? ORDER BY from_id,line,rowid`, id)
	if err != nil {
		return result, Error(err)
	}
	defer rows.Close()
	for rows.Next() {
		var l Backlink
		if err := rows.Scan(&l.From, &l.Line, &l.Alias, &l.Embed); err != nil {
			return result, Error(err)
		}
		result.Backlinks = append(result.Backlinks, l)
	}
	return result, Error(rows.Err())
}

func (i *Index) resolutionNotes(ctx context.Context) ([]kb.Note, error) {
	return resolutionNotesTx(ctx, i.db)
}

func resolutionNotesTx(ctx context.Context, q interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}) ([]kb.Note, error) {
	rows, err := q.QueryContext(ctx, "SELECT id,aliases FROM notes ORDER BY id")
	if err != nil {
		return nil, Error(err)
	}
	defer rows.Close()
	notes := []kb.Note{}
	for rows.Next() {
		var n kb.Note
		var aliases string
		if err := rows.Scan(&n.ID, &aliases); err != nil {
			return nil, Error(err)
		}
		if err := json.Unmarshal([]byte(aliases), &n.Aliases); err != nil {
			return nil, Error(err)
		}
		notes = append(notes, n)
	}
	return notes, Error(rows.Err())
}

// LinkCounts returns the graph totals, independently of report caps.
func (i *Index) LinkCounts(ctx context.Context) (links, broken, orphans int, err error) {
	err = i.db.QueryRowContext(ctx, `SELECT (SELECT count(*) FROM links),(SELECT count(*) FROM links WHERE target_id IS NULL),(SELECT count(*) FROM notes WHERE NOT EXISTS(SELECT 1 FROM links WHERE target_id=notes.id))`).Scan(&links, &broken, &orphans)
	return links, broken, orphans, Error(err)
}
