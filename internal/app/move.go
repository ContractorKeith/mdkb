package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/ContractorKeith/mdkb/internal/index"
	"github.com/ContractorKeith/mdkb/internal/kb"
)

func Move(ctx context.Context, idx *index.Index, root, old, new string, dry bool) (kb.MoveResult, string, error) {
	release, err := kb.LockWrites(root)
	if err != nil {
		return kb.MoveResult{}, "", err
	}
	defer release()
	old, new = kb.CanonicalID(old), kb.CanonicalID(new)
	// Validate before querying so invalid IDs are refused writes.
	if err := kb.PreviewWrite(root, new, "create"); err != nil {
		return kb.MoveResult{}, "", err
	}
	if err := kb.PreviewWrite(root, old, "overwrite"); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return kb.MoveResult{}, "", kb.NoteNotFound(old)
		}
		return kb.MoveResult{}, "", err
	}
	plan, err := kb.PlanMove(root, old, new)
	if err != nil {
		return plan.Result, "", err
	}
	if dry {
		plan.Result.DryRun = true
		var diff strings.Builder

		moved := plan.Changes[0].After
		for _, change := range plan.Changes[1:] {
			if change.ID == old {
				moved = change.After
			}
		}
		diff.WriteString(kb.UnifiedDiff(old+".md", new+".md", plan.Changes[0].Before, moved))
		for _, change := range plan.Changes[1:] {
			if change.ID == old {
				continue
			}
			diff.WriteString(kb.UnifiedDiff(change.ID+".md", change.ID+".md", change.Before, change.After))
		}

		return plan.Result, diff.String(), nil
	}
	if err := kb.ApplyMove(root, plan); err != nil {
		return plan.Result, "", err
	}
	if err := idx.Sync(ctx, root, false); err != nil {
		return plan.Result, "", fmt.Errorf("notes moved but reindex failed: %w", err)
	}
	paths := []string{plan.OldPath, plan.NewPath}
	for _, change := range plan.Changes[1:] {
		if change.ID != old {
			paths = append(paths, change.Path)
		}
	}
	message := fmt.Sprintf("mdkb: mv %s -> %s (%d links rewritten)", old, new, plan.Result.Total)
	committed, commitErr := commitWrite(root, paths, message)
	plan.Result.Commit = committed
	if err := commitErr; err != nil {
		plan.Result.Commit = false
		plan.Result.CommitError = err.Error()
	}
	return plan.Result, "", nil
}
