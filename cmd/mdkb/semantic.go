package main

import (
	"fmt"
	"strings"

	"github.com/ContractorKeith/mdkb/internal/app"
	"github.com/spf13/cobra"
)

func addSemantic(cmd *cobra.Command, root *string, output outputFunc) {
	var model string
	var force, dry bool
	embedCmd := &cobra.Command{Use: "embed", Short: "Embed stale notes with Ollama", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		idx, _, err := synced(cmd, *root, false)
		if err != nil {
			return err
		}
		defer idx.Close()
		s, _, err := app.Searcher(idx, model)
		if err != nil {
			return err
		}
		result, err := app.Embed(cmd.Context(), idx, s.Embedder, force, dry, cmd.ErrOrStderr())
		if err != nil {
			return err
		}
		return output(cmd, result, "Model: %s\nNotes embedded: %d\nChunks: %d\nSkipped: %d\n", result.Model, result.NotesEmbedded, result.Chunks, result.Skipped)
	}}
	embedCmd.Flags().StringVar(&model, "model", "", "Embedding model override")
	embedCmd.Flags().BoolVar(&force, "force", false, "Re-embed all notes")
	embedCmd.Flags().BoolVar(&dry, "dry-run", false, "Preview embedding counts without contacting Ollama")
	cmd.AddCommand(embedCmd)
	var queries, mode string
	evalCmd := &cobra.Command{Use: "eval", Short: "Evaluate fixture retrieval hit@3 and MRR", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		idx, _, err := synced(cmd, *root, false)
		if err != nil {
			return err
		}
		defer idx.Close()
		result, err := app.Eval(cmd.Context(), idx, queries, mode, cmd.ErrOrStderr())
		if err != nil {
			return err
		}
		var text strings.Builder
		fmt.Fprintln(&text, "MODE\tQUERY\tHIT@3\tMRR")
		for _, r := range result.Results {
			for _, q := range r.Queries {
				fmt.Fprintf(&text, "%s\t%s\t%.0f\t%.4f\n", r.Mode, q.Query, q.Hit3, q.MRR)
			}
			fmt.Fprintf(&text, "%s\tMEAN\t%.4f\t%.4f\n", r.Mode, r.Hit3, r.MRR)
		}
		return output(cmd, result, "%s", text.String())
	}}
	evalCmd.Flags().StringVar(&queries, "queries", "", "Relevance JSON file (default: bundled fixture queries)")
	evalCmd.Flags().StringVar(&mode, "mode", "", "keyword, semantic, hybrid (default: all available)")
	cmd.AddCommand(evalCmd)
}
