package index

import (
	"context"
	"github.com/ContractorKeith/mdkb/internal/kb"
)

type Orphan struct {
	ID    string `json:"id" jsonschema:"id"`
	Title string `json:"title" jsonschema:"title"`
}
type Orphans struct {
	Orphans   []Orphan `json:"orphans" jsonschema:"orphans"`
	Total     int      `json:"total" jsonschema:"total"`
	Truncated bool     `json:"truncated" jsonschema:"truncated"`
}

func (i *Index) Orphans(ctx context.Context) (Orphans, error) {
	result := Orphans{Orphans: []Orphan{}}
	filter := ` FROM notes WHERE NOT EXISTS(SELECT 1 FROM links WHERE target_id=notes.id)`
	if err := i.db.QueryRowContext(ctx, "SELECT count(*)"+filter).Scan(&result.Total); err != nil {
		return result, Error(err)
	}
	rows, err := i.db.QueryContext(ctx, "SELECT id,title"+filter+" ORDER BY id LIMIT 50")
	if err != nil {
		return result, Error(err)
	}
	defer rows.Close()
	for rows.Next() {
		var n Orphan
		if err := rows.Scan(&n.ID, &n.Title); err != nil {
			return result, Error(err)
		}
		result.Orphans = append(result.Orphans, n)
	}
	result.Truncated = result.Total > len(result.Orphans)
	return result, Error(rows.Err())
}

type BrokenLink struct {
	From   string `json:"from" jsonschema:"from"`
	Target string `json:"target" jsonschema:"target"`
	Line   int    `json:"line" jsonschema:"line"`
	Reason string `json:"reason" jsonschema:"reason"`
}
type Broken struct {
	Broken    []BrokenLink `json:"broken" jsonschema:"broken"`
	Total     int          `json:"total" jsonschema:"total"`
	Truncated bool         `json:"truncated" jsonschema:"truncated"`
}

func (i *Index) Broken(ctx context.Context) (Broken, error) {
	result := Broken{Broken: []BrokenLink{}}
	notes, err := i.resolutionNotes(ctx)
	if err != nil {
		return result, err
	}
	if err := i.db.QueryRowContext(ctx, "SELECT count(*) FROM links WHERE target_id IS NULL").Scan(&result.Total); err != nil {
		return result, Error(err)
	}
	rows, err := i.db.QueryContext(ctx, "SELECT from_id,target_raw,line FROM links WHERE target_id IS NULL ORDER BY from_id,line,rowid LIMIT 50")
	if err != nil {
		return result, Error(err)
	}
	defer rows.Close()
	resolver := kb.NewResolver(notes)
	for rows.Next() {
		var l BrokenLink
		if err := rows.Scan(&l.From, &l.Target, &l.Line); err != nil {
			return result, Error(err)
		}
		_, l.Reason = resolver.ResolveLink(l.Target)
		result.Broken = append(result.Broken, l)
	}
	result.Truncated = result.Total > len(result.Broken)
	return result, Error(rows.Err())
}
