package index

import (
	"context"
	"encoding/json"
	"strings"
)

type Summary struct {
	Problem string   `json:"problem,omitempty" jsonschema:"problem"`
	ID      string   `json:"id" jsonschema:"id"`
	Title   string   `json:"title" jsonschema:"title"`
	Tags    []string `json:"tags" jsonschema:"tags"`
}
type Listing struct {
	Notes     []Summary `json:"notes" jsonschema:"notes"`
	Total     int       `json:"total" jsonschema:"total"`
	Truncated bool      `json:"truncated" jsonschema:"truncated"`
}
type Tag struct {
	Name  string `json:"name" jsonschema:"name"`
	Count int    `json:"count" jsonschema:"count"`
}

func (i *Index) List(ctx context.Context, tag string, limit int) (Listing, error) {
	result := Listing{Notes: []Summary{}}
	filter := " WHERE (?='' OR EXISTS(SELECT 1 FROM tags WHERE note_id=notes.id AND name=?))"
	if err := i.db.QueryRowContext(ctx, "SELECT count(*) FROM notes"+filter, tag, tag).Scan(&result.Total); err != nil {
		return result, Error(err)
	}
	rows, err := i.db.QueryContext(ctx, "SELECT id,title,tags,problem FROM notes"+filter+" ORDER BY id LIMIT ?", tag, tag, limit)
	if err != nil {
		return result, Error(err)
	}
	defer rows.Close()
	for rows.Next() {
		var n Summary
		var tags string
		if err := rows.Scan(&n.ID, &n.Title, &tags, &n.Problem); err != nil {
			return result, Error(err)
		}
		if err := json.Unmarshal([]byte(tags), &n.Tags); err != nil {
			return result, Error(err)
		}
		result.Notes = append(result.Notes, n)
	}
	result.Truncated = result.Total > len(result.Notes)
	return result, Error(rows.Err())
}
func (i *Index) Tags(ctx context.Context) ([]Tag, error) {
	result := []Tag{}
	rows, err := i.db.QueryContext(ctx, "SELECT name,count(*) FROM tags GROUP BY name ORDER BY count(*) DESC,name")
	if err != nil {
		return nil, Error(err)
	}
	defer rows.Close()
	for rows.Next() {
		var t Tag
		if err := rows.Scan(&t.Name, &t.Count); err != nil {
			return nil, Error(err)
		}
		result = append(result, t)
	}
	return result, Error(rows.Err())
}

// Match is the storage result; search owns the public Hit contract.
type Match struct {
	ID, Title string
	Score     float64
	Snippet   string
}

func (i *Index) Keyword(ctx context.Context, query string, limit int) ([]Match, error) {
	result, err := i.matches(ctx, query, limit)
	if err != nil && isSyntaxError(err.Error()) {
		return i.matches(ctx, `"`+strings.ReplaceAll(query, `"`, `""`)+`"`, limit)
	}
	return result, err
}
func isSyntaxError(message string) bool {
	for _, part := range []string{"fts5: syntax error", "unterminated string", "no such column", "unknown special query", "fts5: phrase queries"} {
		if strings.Contains(message, part) {
			return true
		}
	}
	return false
}
func (i *Index) matches(ctx context.Context, q string, limit int) ([]Match, error) {
	result := []Match{}
	rows, err := i.db.QueryContext(ctx, `SELECT notes.id,notes.title,bm25(notes_fts),snippet(notes_fts,-1,'[',']',' … ',24) FROM notes_fts JOIN notes ON notes.rowid=notes_fts.rowid WHERE notes_fts MATCH ? ORDER BY bm25(notes_fts),notes.id LIMIT ?`, q, limit)
	if err != nil {
		return nil, Error(err)
	}
	defer rows.Close()
	for rows.Next() {
		var m Match
		if err := rows.Scan(&m.ID, &m.Title, &m.Score, &m.Snippet); err != nil {
			return nil, Error(err)
		}
		result = append(result, m)
	}
	return result, Error(rows.Err())
}
