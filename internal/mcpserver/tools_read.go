package mcpserver

import (
	"context"
	"fmt"

	"github.com/ContractorKeith/mdkb/internal/app"
	"github.com/ContractorKeith/mdkb/internal/index"
	"github.com/ContractorKeith/mdkb/internal/kb"
	"github.com/ContractorKeith/mdkb/internal/search"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type listInput struct {
	Tag   string `json:"tag,omitempty" jsonschema:"Exact tag filter"`
	Limit *int   `json:"limit,omitempty" jsonschema:"Positive result limit capped at 50; default 50"`
}
type idInput struct {
	ID string `json:"id" jsonschema:"Relative note path without .md using forward slashes"`
}
type searchInput struct {
	Query string `json:"query" jsonschema:"FTS5 keyword query; malformed expressions retry as a literal phrase"`
	Limit *int   `json:"limit,omitempty" jsonschema:"Positive result limit capped at 50; default 50"`
}
type readOutput struct {
	ID      string   `json:"id" jsonschema:"Note ID"`
	Title   string   `json:"title" jsonschema:"Display title"`
	Tags    []string `json:"tags" jsonschema:"Tags"`
	Aliases []string `json:"aliases" jsonschema:"Link aliases"`
	Body    string   `json:"body" jsonschema:"Markdown body excluding valid frontmatter"`
	Problem string   `json:"problem,omitempty" jsonschema:"Frontmatter parse problem if present"`
}
type searchOutput struct {
	Hits []search.Hit `json:"hits" jsonschema:"Ranked keyword matches"`
}
type emptyInput struct{}

func limit(value *int) (int, error) {
	if value == nil {
		return 50, nil
	}
	if *value < 1 {
		return 0, fmt.Errorf("limit must be positive")
	}
	return min(*value, 50), nil
}

func addReads(s *mcp.Server, state *server) {
	register(s, state, "list_notes", "List notes by relative path IDs without .md, capped at 50 with total and truncated; invalid limits or index failures return tool errors.", func(ctx context.Context, idx *index.Index, in listInput) (index.Listing, error) {
		n, err := limit(in.Limit)
		if err != nil {
			return index.Listing{}, err
		}
		return idx.List(ctx, in.Tag, n)
	})
	register(s, state, "read_note", "Read one relative path ID without .md including its body and any parse problem; unknown or unsafe IDs and index failures return tool errors.", func(ctx context.Context, _ *index.Index, in idInput) (readOutput, error) {
		n, _, err := kb.Read(state.root, in.ID)
		return readOutput{n.ID, n.Title, n.Tags, n.Aliases, n.Body, n.Problem}, err
	})
	register(s, state, "search_notes", "Search keywords and return relative path IDs without .md capped at 50; invalid limits or index failures return tool errors and no matches returns an empty list.", func(ctx context.Context, idx *index.Index, in searchInput) (searchOutput, error) {
		n, err := limit(in.Limit)
		if err != nil {
			return searchOutput{}, err
		}
		hits, err := (search.Search{Index: idx}).Keyword(ctx, in.Query, n)
		return searchOutput{hits}, err
	})
	register(s, state, "get_links", "Return all outgoing links for one relative path ID without .md in body order; unknown IDs or index failures return tool errors.", func(ctx context.Context, idx *index.Index, in idInput) (index.Outgoing, error) {
		return idx.Links(ctx, in.ID)
	})
	register(s, state, "get_backlinks", "Return all incoming links for one relative path ID without .md in source order; unknown IDs or index failures return tool errors.", func(ctx context.Context, idx *index.Index, in idInput) (index.Incoming, error) {
		return idx.Backlinks(ctx, in.ID)
	})
	register(s, state, "get_info", "Return root, cache and uncapped note and link totals for relative path IDs without .md; initialization or index failures return tool errors.", func(ctx context.Context, _ *index.Index, _ emptyInput) (kb.Details, error) {
		return app.Info(ctx, state.root)
	})
}
