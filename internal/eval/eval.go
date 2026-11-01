// Package eval measures fixture retrieval against explicit relevance judgments.
package eval

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/ContractorKeith/mdkb/internal/search"
	evaldata "github.com/ContractorKeith/mdkb/testdata/eval"
)

type Query struct {
	Query    string   `json:"query"`
	Expected []string `json:"expected"`
}
type Row struct {
	Query string   `json:"query"`
	Hit3  float64  `json:"hit_at_3"`
	MRR   float64  `json:"mrr"`
	IDs   []string `json:"ids"`
}
type Result struct {
	Mode    string  `json:"mode"`
	Queries []Row   `json:"queries"`
	Hit3    float64 `json:"hit_at_3"`
	MRR     float64 `json:"mrr"`
}

func Load(path string) ([]Query, error) {
	raw := evaldata.Queries
	var err error
	if path != "" {
		raw, err = os.ReadFile(path)
		if err != nil {
			return nil, err
		}
	}
	var queries []Query
	if err = json.Unmarshal(raw, &queries); err != nil {
		return nil, err
	}
	if len(queries) == 0 {
		return nil, fmt.Errorf("queries must not be empty")
	}
	for _, q := range queries {
		if q.Query == "" || len(q.Expected) == 0 {
			return nil, fmt.Errorf("queries require text and expected IDs")
		}
	}
	return queries, nil
}
func Run(ctx context.Context, s search.Search, queries []Query, mode string) (Result, error) {
	result := Result{Mode: mode, Queries: []Row{}}
	for _, q := range queries {
		var hits []search.Hit
		var err error
		switch mode {
		case "keyword":
			hits, err = s.Keyword(ctx, q.Query, 50)
		case "semantic":
			hits, err = s.Semantic(ctx, q.Query, 50)
		case "hybrid":
			hits, err = s.Hybrid(ctx, q.Query, 50)
		default:
			return result, fmt.Errorf("unknown eval mode %q", mode)
		}
		if err != nil {
			return result, err
		}
		row := Row{Query: q.Query, IDs: []string{}}
		for rank, h := range hits {
			if rank < 3 {
				row.IDs = append(row.IDs, h.ID)
			}
			for _, id := range q.Expected {
				if h.ID == id {
					if rank < 3 {
						row.Hit3 = 1
					}
					if row.MRR == 0 {
						row.MRR = 1 / float64(rank+1)
					}
					break
				}
			}
		}
		result.Queries = append(result.Queries, row)
		result.Hit3 += row.Hit3
		result.MRR += row.MRR
	}
	result.Hit3 /= float64(len(queries))
	result.MRR /= float64(len(queries))
	return result, nil
}
