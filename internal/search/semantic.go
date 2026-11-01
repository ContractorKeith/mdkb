package search

import (
	"context"
	"fmt"
	"math"
	"sort"

	"github.com/ContractorKeith/mdkb/internal/embed"
)

func (s Search) Semantic(ctx context.Context, q string, limit int) ([]Hit, error) {
	vectors, err := s.Index.Vectors(ctx, s.Embedder.Model())
	if err != nil {
		return nil, err
	}
	if len(vectors) == 0 {
		return nil, &embed.ErrUnavailable{Model: s.Embedder.Model(), Reason: "no embeddings; run mdkb embed"}
	}
	query, err := s.Embedder.Embed(ctx, []string{q})
	if err != nil {
		return nil, err
	}
	if len(query) != 1 {
		return nil, fmt.Errorf("expected one query embedding")
	}
	best := map[string]Hit{}
	chunks := map[string][]string{}
	for _, v := range vectors {
		if len(v.Values) != len(query[0]) {
			return nil, fmt.Errorf("embedding dimensions differ; run mdkb embed --force")
		}
		score := cosine(query[0], v.Values)
		old, ok := best[v.ID]
		if ok && score <= old.Score {
			continue
		}
		if chunks[v.ID] == nil {
			chunks[v.ID] = embed.Chunks(v.Title, v.Body)
		}
		if v.Chunk < 0 || v.Chunk >= len(chunks[v.ID]) {
			return nil, fmt.Errorf("invalid stored chunk")
		}
		snippet := []rune(chunks[v.ID][v.Chunk])
		snippet = snippet[:min(200, len(snippet))]
		best[v.ID] = Hit{ID: v.ID, Title: v.Title, Score: score, Snippet: string(snippet)}
	}
	hits := []Hit{}
	for _, h := range best {
		hits = append(hits, h)
	}
	rank(hits)
	hits = hits[:min(limit, len(hits))]
	for j := range hits {
		hits[j].Score = round(hits[j].Score)
	}
	return hits, nil
}
func cosine(a, b []float32) float64 {
	var dot, aa, bb float64
	for j, x := range a {
		y := float64(b[j])
		xx := float64(x)
		dot += xx * y
		aa += xx * xx
		bb += y * y
	}
	if aa == 0 || bb == 0 {
		return 0
	}
	return dot / math.Sqrt(aa*bb)
}
func round(x float64) float64 { return math.Round(x*1e4) / 1e4 }
func rank(h []Hit) {
	sort.Slice(h, func(i, j int) bool {
		if h[i].Score == h[j].Score {
			return h[i].ID < h[j].ID
		}
		return h[i].Score > h[j].Score
	})
}

func (s Search) Hybrid(ctx context.Context, q string, limit int) ([]Hit, error) {
	keyword, err := s.Keyword(ctx, q, 50)
	if err != nil {
		return nil, err
	}
	semantic, err := s.Semantic(ctx, q, 50)
	if err != nil {
		return nil, err
	}
	return Fuse(keyword, semantic, limit), nil
}

// Fuse uses one-based ranks and sorts before rounding to preserve close scores.
func Fuse(keyword, semantic []Hit, limit int) []Hit {
	merged := map[string]Hit{}
	for j, list := range [][]Hit{keyword, semantic} {
		source := []string{"keyword", "semantic"}[j]
		for r, h := range list {
			old, ok := merged[h.ID]
			if !ok {
				old = h
				old.Score = 0
				old.Sources = nil
			}
			old.Score += 1 / float64(60+r+1)
			old.Sources = append(old.Sources, source)
			merged[h.ID] = old
		}
	}
	hits := []Hit{}
	for _, h := range merged {
		hits = append(hits, h)
	}
	rank(hits)
	hits = hits[:min(limit, len(hits))]
	for j := range hits {
		hits[j].Score = round(hits[j].Score)
	}
	return hits
}
