package mcpserver

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ContractorKeith/mdkb/internal/app"
	"github.com/ContractorKeith/mdkb/internal/index"
	"github.com/ContractorKeith/mdkb/internal/kb"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type writeInput struct {
	ID          string          `json:"id" jsonschema:"Relative note path without .md; no dot segments or symlinks"`
	Body        string          `json:"body" jsonschema:"Markdown body to write"`
	Mode        string          `json:"mode" jsonschema:"Required mode: create, append, or overwrite"`
	Frontmatter *kb.Frontmatter `json:"frontmatter,omitempty" jsonschema:"Optional metadata for create or overwrite; refused on append"`
}
type captureInput struct {
	Text  string `json:"text" jsonschema:"Markdown text to capture"`
	Title string `json:"title,omitempty" jsonschema:"Optional title and slug source"`
}

func addWrites(s *mcp.Server, state *server) {
	register(s, state, "write_note", "Create, append to or overwrite a relative path ID without .md; unsafe IDs, existing creates, missing append/overwrite targets or index failures return tool errors.", func(ctx context.Context, idx *index.Index, in writeInput) (kb.WriteResult, error) {
		if in.Mode == "append" && in.Frontmatter != nil {
			return kb.WriteResult{}, fmt.Errorf("frontmatter cannot be appended")
		}
		content, err := kb.Content(in.Body, in.Frontmatter)
		if err != nil {
			return kb.WriteResult{}, err
		}
		return writeAndSync(ctx, idx, state.root, in.ID, content, in.Mode)
	})
	register(s, state, "capture", "Create a timestamped relative path ID without .md in 00-Inbox with a slug capped at 40 characters; collisions, write or index failures return tool errors.", func(ctx context.Context, idx *index.Index, in captureInput) (kb.WriteResult, error) {
		id, content, err := kb.CaptureContent(in.Text, in.Title, time.Now())
		if err != nil {
			return kb.WriteResult{}, err
		}
		result, err := app.Write(ctx, idx, state.root, id, content, "create", "capture")
		if result.CommitError != "" {
			log.Print(result.CommitError)
		}
		return result, err
	})
}

func writeAndSync(ctx context.Context, idx *index.Index, root, id, content, mode string) (kb.WriteResult, error) {
	result, err := app.Write(ctx, idx, root, id, content, mode, mode)
	if result.CommitError != "" {
		log.Print(result.CommitError)
	}
	return result, err
}
