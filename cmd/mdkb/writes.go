package main

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/ContractorKeith/mdkb/internal/app"
	"github.com/ContractorKeith/mdkb/internal/config"
	"github.com/ContractorKeith/mdkb/internal/kb"
	"github.com/spf13/cobra"
)

type writeFlags struct {
	body, title, tags, dir, date string
	stdin, dry                   bool
}

func writeRoot(root string) (string, error) {
	path, err := config.ResolveRoot(root)
	if err != nil {
		return "", err
	}
	info, err := kb.Info(path)
	if err != nil || !info.Initialized {
		return "", fmt.Errorf("%w: KB not initialized: %s", kb.ErrWriteRefused, path)
	}
	return path, nil
}

func (f *writeFlags) text(cmd *cobra.Command) (string, error) {
	if f.stdin {
		raw, err := io.ReadAll(cmd.InOrStdin())
		return string(raw), err
	}
	return f.body, nil
}

func publish(cmd *cobra.Command, root, id, content, mode, verb string, dry bool, output outputFunc) error {
	if err := kb.PreviewWrite(root, id, mode); err != nil {
		return err
	}
	if dry {
		if mode == "append" {
			_, raw, err := kb.Read(root, id)
			if err != nil {
				return err
			}
			content = strings.TrimSuffix(raw, "\n") + "\n\n" + content
		}
		fmt.Fprintln(cmd.ErrOrStderr(), strings.TrimSuffix(content, "\n"))
		return output(cmd, struct {
			ID     string `json:"id"`
			Path   string `json:"path"`
			DryRun bool   `json:"dry_run"`
		}{id, id + ".md", true}, "%s\t%s\n", id, id+".md")
	}
	idx, _, err := app.Synced(cmd.Context(), root, false)
	if err != nil {
		return err
	}
	defer idx.Close()
	result, err := app.Write(cmd.Context(), idx, root, id, content, mode, verb)
	if err != nil {
		return err
	}
	if result.CommitError != "" {
		fmt.Fprintln(cmd.ErrOrStderr(), "mdkb:", result.CommitError)
	}
	return output(cmd, result, "%s\t%s\n", result.ID, result.Path)
}

func addWrites(parent *cobra.Command, root *string, output outputFunc) {
	var n writeFlags
	newCmd := &cobra.Command{Use: "new <title>", Short: "Create a Markdown note", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		path, err := writeRoot(*root)
		if err != nil {
			return err
		}
		body, err := n.text(cmd)
		if err != nil {
			return err
		}
		id := kb.Slug(args[0])
		if n.dir != "" {
			id = n.dir + "/" + id
		}
		tags := []string{}
		if n.tags != "" {
			tags = strings.Split(n.tags, ",")
		}
		content, err := kb.Content(body, &kb.Frontmatter{Title: args[0], Tags: tags, Created: time.Now().Format("2006-01-02")})
		if err != nil {
			return err
		}
		return publish(cmd, path, id, content, "create", "new", n.dry, output)
	}}
	newCmd.Flags().StringVar(&n.body, "body", "", "Markdown body")
	newCmd.Flags().StringVar(&n.tags, "tags", "", "Comma-separated tags")
	newCmd.Flags().StringVar(&n.dir, "dir", "", "Relative note directory")
	newCmd.Flags().BoolVar(&n.stdin, "stdin", false, "Read body from stdin")
	newCmd.Flags().BoolVar(&n.dry, "dry-run", false, "Preview without writing")
	newCmd.MarkFlagsMutuallyExclusive("body", "stdin")
	parent.AddCommand(newCmd)
	var c writeFlags
	capture := &cobra.Command{Use: "capture [TEXT...]", Short: "Capture a note in 00-Inbox", RunE: func(cmd *cobra.Command, args []string) error {
		if c.stdin && len(args) > 0 {
			return errors.New("TEXT and --stdin are mutually exclusive")
		}
		path, err := writeRoot(*root)
		if err != nil {
			return err
		}
		c.body = strings.Join(args, " ")
		text, err := c.text(cmd)
		if err != nil {
			return err
		}
		id, content, err := kb.CaptureContent(text, c.title, time.Now())
		if err != nil {
			return err
		}
		return publish(cmd, path, id, content, "create", "capture", c.dry, output)
	}}
	capture.Flags().StringVar(&c.title, "title", "", "Capture title")
	capture.Flags().BoolVar(&c.stdin, "stdin", false, "Read text from stdin")
	capture.Flags().BoolVar(&c.dry, "dry-run", false, "Preview without writing")
	parent.AddCommand(capture)
	addDaily(parent, root, output)
	addMove(parent, root, output)
}
