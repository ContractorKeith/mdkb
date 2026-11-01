package index

import (
	"context"
	"database/sql"

	"github.com/ContractorKeith/mdkb/internal/kb"
)

func replaceLinks(ctx context.Context, tx *sql.Tx, n kb.Note) error {
	if _, err := tx.ExecContext(ctx, "DELETE FROM links WHERE from_id=?", n.ID); err != nil {
		return err
	}
	for _, ref := range kb.ExtractLinks(n.Body) {
		kind := "link"
		if ref.Embed {
			kind = "embed"
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO links(from_id,target_raw,kind,heading,alias,line) VALUES(?,?,?,?,?,?)`, n.ID, ref.Target, kind, ref.Heading, ref.Alias, ref.Line); err != nil {
			return err
		}
	}
	_, err := tx.ExecContext(ctx, "INSERT OR IGNORE INTO link_sources(note_id) VALUES(?)", n.ID)
	return err
}

// Reconsider all targets: additions can make old basename/alias matches ambiguous.
func resolveLinks(ctx context.Context, tx *sql.Tx, notes []kb.Note) error {
	rows, err := tx.QueryContext(ctx, "SELECT DISTINCT target_raw FROM links")
	if err != nil {
		return err
	}
	targets := []string{}
	for rows.Next() {
		var target string
		if err := rows.Scan(&target); err != nil {
			rows.Close()
			return err
		}
		targets = append(targets, target)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	resolver := kb.NewResolver(notes)
	for _, target := range targets {
		id, reason := resolver.ResolveLink(target)
		var resolved any
		if reason == "" {
			resolved = id
		}
		if _, err := tx.ExecContext(ctx, "UPDATE links SET target_id=? WHERE target_raw=? AND target_id IS NOT ?", resolved, target, resolved); err != nil {
			return err
		}
	}
	return nil
}
