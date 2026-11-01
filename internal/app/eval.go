package app

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/ContractorKeith/mdkb/internal/embed"
	"github.com/ContractorKeith/mdkb/internal/eval"
	"github.com/ContractorKeith/mdkb/internal/index"
)

type EvalResult struct {
	Results []eval.Result `json:"results"`
}

func Eval(ctx context.Context, idx *index.Index, path, mode string, diagnostics io.Writer) (EvalResult, error) {
	result := EvalResult{Results: []eval.Result{}}
	queries, err := eval.Load(path)
	if err != nil {
		return result, err
	}
	s, _, err := Searcher(idx, "")
	if err != nil {
		return result, err
	}
	modes := []string{mode}
	if mode == "" {
		modes = []string{"keyword"}
		counts, err := idx.EmbeddingCounts(ctx, s.Embedder.Model())
		if err != nil {
			return result, err
		}
		if counts.Chunks > 0 {
			modes = append(modes, "semantic", "hybrid")
		}
	}
	for _, m := range modes {
		r, err := eval.Run(ctx, s, queries, m)
		var unavailable *embed.ErrUnavailable
		if mode == "" && errors.As(err, &unavailable) {
			if diagnostics != nil {
				fmt.Fprintf(diagnostics, "semantic unavailable (%s); keyword only\n", unavailable)
			}
			break
		}
		if err != nil {
			return result, err
		}
		result.Results = append(result.Results, r)
	}
	return result, nil
}
