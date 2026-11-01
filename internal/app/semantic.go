package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/ContractorKeith/mdkb/internal/config"
	"github.com/ContractorKeith/mdkb/internal/embed"
	"github.com/ContractorKeith/mdkb/internal/index"
	"github.com/ContractorKeith/mdkb/internal/search"
)

func Searcher(idx *index.Index, model string) (search.Search, bool, error) {
	cfg, err := config.Semantic()
	if model != "" {
		cfg.Model = model
	}
	return search.Search{Index: idx, Embedder: embed.Ollama{URL: cfg.URL, ModelName: cfg.Model, Client: &http.Client{Timeout: time.Duration(cfg.TimeoutSeconds) * time.Second}}}, cfg.Enabled, err
}

type SearchResult struct {
	Mode string       `json:"mode"`
	Hits []search.Hit `json:"hits"`
}

// searchMode resolves automatic selection while preserving explicit mode failures.
func searchMode(ctx context.Context, idx *index.Index, q, mode string, limit int) (SearchResult, error) {
	result := SearchResult{Mode: "keyword", Hits: []search.Hit{}}
	if limit < 1 {
		return result, fmt.Errorf("limit must be positive")
	}
	if mode != "" && mode != "keyword" && mode != "semantic" && mode != "hybrid" {
		return result, fmt.Errorf("unknown search mode %q", mode)
	}
	s, enabled, err := Searcher(idx, "")
	if err != nil {
		return result, err
	}
	if mode == "" {
		counts, err := idx.EmbeddingCounts(ctx, s.Embedder.Model())
		if err != nil {
			return result, err
		}
		mode = "keyword"
		if enabled && counts.Chunks > 0 {
			mode = "hybrid"
		}
	}
	result.Mode = mode
	switch mode {
	case "keyword":
		result.Hits, err = s.Keyword(ctx, q, limit)
	case "semantic":
		result.Hits, err = s.Semantic(ctx, q, limit)
	case "hybrid":
		result.Hits, err = s.Hybrid(ctx, q, limit)
	}
	return result, err
}

// SearchDefault permits fallback only for an automatically selected hybrid mode.
func SearchDefault(ctx context.Context, idx *index.Index, q, mode string, limit int, diagnostics io.Writer) (SearchResult, error) {
	result, err := searchMode(ctx, idx, q, mode, limit)
	var unavailable *embed.ErrUnavailable
	if mode == "" && result.Mode == "hybrid" && errors.As(err, &unavailable) {
		if diagnostics == nil {
			diagnostics = os.Stderr
		}
		fmt.Fprintf(diagnostics, "semantic unavailable (%s); keyword only\n", unavailable)
		return searchMode(ctx, idx, q, "keyword", limit)
	}
	return result, err
}
