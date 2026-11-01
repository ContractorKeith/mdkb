package app

import (
	"context"
	"fmt"

	"github.com/ContractorKeith/mdkb/internal/index"
	"github.com/ContractorKeith/mdkb/internal/kb"
)

// Write publishes, reindexes and commits a logical write for both front ends.
func Write(ctx context.Context, idx *index.Index, root, id, content, mode, verb string) (kb.WriteResult, error) {
	result, err := kb.Write(root, id, content, mode)
	if err != nil {
		return result, err
	}
	if err := idx.SyncNote(ctx, root, id); err != nil {
		return result, fmt.Errorf("note written but reindex failed: %w", err)
	}
	if !result.Changed {
		return result, nil
	}
	if verb == "create" {
		verb = "new"
	}
	committed, commitErr := kb.CommitWriteStatus(root, []string{result.Path}, "mdkb: "+verb+" "+id)
	result.Commit = committed
	if err := commitErr; err != nil {
		result.Commit = false
		result.CommitError = err.Error()
	}
	return result, nil
}
