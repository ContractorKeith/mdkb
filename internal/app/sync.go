// Package app shares synchronized operations between the CLI and MCP front ends.
package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/ContractorKeith/mdkb/internal/config"
	"github.com/ContractorKeith/mdkb/internal/index"
	"github.com/ContractorKeith/mdkb/internal/kb"
)

var ErrUninitialized = errors.New("KB not initialized")

// Synced resolves the root and opens an incrementally synchronized cache.
func Synced(ctx context.Context, root string, full bool) (*index.Index, string, error) {
	path, err := config.ResolveRoot(root)
	if err != nil {
		return nil, "", err
	}
	info, err := kb.Info(path)
	if err != nil || !info.Initialized {
		return nil, "", fmt.Errorf("%w: %s", ErrUninitialized, path)
	}
	// Initial sync also writes SQLite; serialize it with note publication.
	release, err := kb.LockWrites(path)
	if err != nil {
		return nil, "", err
	}
	defer release()
	var idx *index.Index
	if full {
		idx, err = index.OpenForReindex(path)
	} else {
		idx, err = index.Open(path)
	}
	if err != nil {
		return nil, "", err
	}
	if err := idx.Sync(ctx, path, full); err != nil {
		idx.Close()
		return nil, "", err
	}
	return idx, path, nil
}

// Info fills the same inspector response for both front ends.
func Info(ctx context.Context, root string) (kb.Details, error) {
	path, err := config.ResolveRoot(root)
	if err != nil {
		return kb.Details{}, err
	}
	info, err := kb.Info(path)
	cfg, cfgErr := config.Semantic()
	if cfgErr != nil {
		return info, cfgErr
	}
	info.Embeddings.Model = cfg.Model
	if err != nil || !info.Initialized {
		return info, err
	}
	info.IndexPath, err = index.CachePath(path)
	if err != nil {
		return info, err
	}
	idx, _, err := Synced(ctx, path, false)
	if err != nil {
		if index.IsUnreadable(err) {
			info.Problem = map[string]string{"index": "unreadable"}
			return info, nil
		}
		return info, err
	}
	defer idx.Close()
	info.IndexedNotes, err = idx.Count(ctx)
	if err != nil {
		return info, err
	}
	info.IndexPath = idx.Path
	info.Links, info.BrokenLinks, info.Orphans, err = idx.LinkCounts(ctx)
	if err == nil {
		counts, e := idx.EmbeddingCounts(ctx, cfg.Model)
		err = e
		info.Embeddings.Model = counts.Model
		info.Embeddings.Notes = counts.Notes
		info.Embeddings.Chunks = counts.Chunks
	}
	return info, err
}
