// Package search provides ranked search results over the index.
package search

import (
	"context"

	"github.com/ContractorKeith/mdkb/internal/index"
)

type Hit struct {
	ID      string  `json:"id"`
	Title   string  `json:"title"`
	Score   float64 `json:"score"`
	Snippet string  `json:"snippet"`
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
		hits = append(hits, Hit{m.ID, m.Title, m.Score, m.Snippet})
	}
	return hits, nil
}
