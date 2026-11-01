package app

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/ContractorKeith/mdkb/internal/embed"
	"github.com/ContractorKeith/mdkb/internal/index"
)

type EmbedResult struct {
	Model         string `json:"model"`
	NotesEmbedded int    `json:"notes_embedded"`
	Chunks        int    `json:"chunks"`
	Skipped       int    `json:"skipped"`
	DurationMS    int64  `json:"duration_ms"`
	DryRun        bool   `json:"dry_run,omitempty"`
}

func Embed(ctx context.Context, idx *index.Index, e embed.Embedder, force, dry bool, progress io.Writer) (EmbedResult, error) {
	start := time.Now()
	result := EmbedResult{Model: e.Model(), DryRun: dry}
	notes, err := idx.EmbedNotes(ctx, e.Model())
	if err != nil {
		return result, err
	}
	for _, n := range notes {
		if n.Current && !force {
			result.Skipped++
			continue
		}
		chunks := embed.Chunks(n.Title, n.Body)
		if !dry {
			vectors, err := e.Embed(ctx, chunks)
			if err != nil {
				return result, err
			}
			if len(vectors) != len(chunks) {
				return result, fmt.Errorf("embedding count mismatch")
			}
			if err = idx.StoreEmbeddings(ctx, n, e.Model(), vectors); err != nil {
				return result, err
			}
		}
		result.NotesEmbedded++
		result.Chunks += len(chunks)
		if result.NotesEmbedded%25 == 0 && progress != nil {
			fmt.Fprintf(progress, "Embedded %d notes\n", result.NotesEmbedded)
		}
	}
	result.DurationMS = time.Since(start).Milliseconds()
	return result, nil
}
