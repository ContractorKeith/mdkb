package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/ContractorKeith/mdkb/internal/app"
	"github.com/ContractorKeith/mdkb/internal/config"
	"github.com/ContractorKeith/mdkb/internal/kb"
	"github.com/spf13/cobra"
)

func addDaily(parent *cobra.Command, root *string, output outputFunc) {
	var f writeFlags
	cmd := &cobra.Command{Use: "daily", Short: "Open or append to a daily note", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, args []string) error {
		path, err := writeRoot(*root)
		if err != nil {
			return err
		}
		date := f.date
		if date == "" {
			date = time.Now().Format("2006-01-02")
		}
		if _, err := time.Parse("2006-01-02", date); err != nil {
			return fmt.Errorf("invalid date: %w", err)
		}
		cfg, err := config.Writes()
		if err != nil {
			return err
		}
		id := cfg.DailyDir + "/" + date
		// Validate the configured directory before treating a missing file as create.
		if err := kb.ValidateID(id, true); err != nil {
			return err
		}
		_, _, readErr := kb.Read(path, id)
		body, err := f.text(cmd)
		if err != nil {
			return err
		}
		mode := "append"
		if errors.Is(readErr, os.ErrNotExist) {
			mode = "create"
			body, err = kb.Content("# "+date+"\n\n"+body, &kb.Frontmatter{Created: date})
			if err != nil {
				return err
			}
		} else if readErr != nil {
			return readErr
		} else if !f.stdin && !cmd.Flags().Changed("append") {
			return output(cmd, struct {
				ID   string `json:"id"`
				Path string `json:"path"`
			}{id, id + ".md"}, "%s\t%s\n", id, id+".md")
		}
		return publish(cmd, path, id, body, mode, "daily", f.dry, output)
	}}
	cmd.Flags().StringVar(&f.body, "append", "", "Append Markdown text")
	cmd.Flags().StringVar(&f.date, "date", "", "Daily date (YYYY-MM-DD)")
	cmd.Flags().BoolVar(&f.stdin, "stdin", false, "Read appended text from stdin")
	cmd.Flags().BoolVar(&f.dry, "dry-run", false, "Preview without writing")
	cmd.MarkFlagsMutuallyExclusive("append", "stdin")
	parent.AddCommand(cmd)
}

func addMove(parent *cobra.Command, root *string, output outputFunc) {
	var dry bool
	cmd := &cobra.Command{Use: "mv <old> <new>", Short: "Move a note and rewrite inbound links", Args: cobra.ExactArgs(2), RunE: func(cmd *cobra.Command, args []string) error {
		path, err := writeRoot(*root)
		if err != nil {
			return err
		}
		for _, id := range args {
			if err := kb.ValidateID(id, true); err != nil {
				return err
			}
		}
		idx, _, err := synced(cmd, path, false)
		if err != nil {
			return err
		}
		defer idx.Close()
		result, diff, err := app.Move(cmd.Context(), idx, path, args[0], args[1], dry)
		if err != nil {
			return err
		}
		if diff != "" {
			fmt.Fprint(cmd.ErrOrStderr(), diff)
		}
		if result.CommitError != "" {
			fmt.Fprintln(cmd.ErrOrStderr(), "mdkb:", result.CommitError)
		}
		return output(cmd, result, "%s -> %s (%d links rewritten)\n", result.Old, result.New, result.Total)
	}}
	cmd.Flags().BoolVar(&dry, "dry-run", false, "Preview JSON and unified diffs")
	parent.AddCommand(cmd)
}
