// Package mcpserver exposes the knowledge base over MCP stdio.
package mcpserver

import (
	"context"
	"sync"

	"github.com/ContractorKeith/mdkb/internal/app"
	"github.com/ContractorKeith/mdkb/internal/index"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// New checks initialization before registering tools.
func New(root string, allowWrite bool) (*mcp.Server, error) {
	idx, path, err := app.Synced(context.Background(), root, false)
	if err != nil {
		return nil, err
	}
	idx.Close()
	s := mcp.NewServer(&mcp.Implementation{Name: "mdkb", Version: "dev"}, nil)
	state := &server{root: path}
	addReads(s, state)
	if allowWrite {
		addWrites(s, state)
	}
	return s, nil
}

// Run gives stdout exclusively to the stdio transport.
func Run(ctx context.Context, root string, allowWrite bool) error {
	s, err := New(root, allowWrite)
	if err != nil {
		return err
	}
	return s.Run(ctx, &mcp.StdioTransport{})
}

type server struct {
	root string
	mu   sync.Mutex
}

// register serializes sync/query/write operations across this server's sessions.
func register[In, Out any](s *mcp.Server, state *server, name, description string, fn func(context.Context, *index.Index, In) (Out, error)) {
	mcp.AddTool(s, &mcp.Tool{Name: name, Description: description}, func(ctx context.Context, _ *mcp.CallToolRequest, in In) (*mcp.CallToolResult, Out, error) {
		state.mu.Lock()
		defer state.mu.Unlock()
		var zero Out
		idx, _, err := app.Synced(ctx, state.root, false)
		if err != nil {
			return nil, zero, err
		}
		defer idx.Close()
		out, err := fn(ctx, idx, in)
		return nil, out, err
	})
}
