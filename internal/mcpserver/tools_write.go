package mcpserver

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode"

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
		now := time.Now()
		source := in.Title
		if source == "" {
			source = in.Text
		}
		id := "00-Inbox/" + now.Format("2006-01-02-150405") + "-" + slug(source)
		content, err := kb.Content(in.Text, &kb.Frontmatter{Title: in.Title, Created: now.Format(time.RFC3339)})
		if err != nil {
			return kb.WriteResult{}, err
		}
		return writeAndSync(ctx, idx, state.root, id, content, "create")
	})
}

func writeAndSync(ctx context.Context, idx *index.Index, root, id, content, mode string) (kb.WriteResult, error) {
	result, err := kb.Write(root, id, content, mode)
	if err != nil {
		return result, err
	}
	if err := idx.SyncNote(ctx, root, id); err != nil {
		return result, fmt.Errorf("note written but reindex failed: %w", err)
	}
	return result, nil
}

func slug(text string) string {
	var result []rune
	for _, r := range strings.ToLower(text) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			result = append(result, r)
		} else if len(result) > 0 && result[len(result)-1] != '-' {
			result = append(result, '-')
		}
		if len(result) == 40 {
			break
		}
	}
	value := strings.Trim(string(result), "-")
	if value == "" {
		return "note"
	}
	return value
}
