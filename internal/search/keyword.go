// Package search provides ranked search results over the index.
package search

import (
	"context"
	"math"

	"github.com/ContractorKeith/mdkb/internal/index"
)

type Hit struct {
	ID      string  `json:"id" jsonschema:"id"`
	Title   string  `json:"title" jsonschema:"title"`
	Score   float64 `json:"score" jsonschema:"score"`
	Snippet string  `json:"snippet" jsonschema:"snippet"`
}
type Search struct{ Index *index.Index }

// Keyword returns BM25 results, with lower scores ranking first.
func (s Search) Keyword(ctx context.Context, q string, limit int) ([]Hit, error) {
	matches, err := s.Index.Keyword(ctx, q, limit)
	if err != nil {
		return nil, err
	}
	hits := []Hit{}
	for _, m := range matches {
		hits = append(hits, Hit{m.ID, m.Title, math.Round(m.Score*1e4) / 1e4, m.Snippet})
	}
	return hits, nil
}
