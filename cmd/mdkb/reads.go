package main

import (
	"fmt"
	"strings"

	"github.com/ContractorKeith/mdkb/internal/app"
	"github.com/ContractorKeith/mdkb/internal/index"
	"github.com/ContractorKeith/mdkb/internal/kb"
	"github.com/spf13/cobra"
)

var errUninitialized = app.ErrUninitialized

type outputFunc func(*cobra.Command, any, string, ...any) error

func synced(cmd *cobra.Command, root string, full bool) (*index.Index, string, error) {
	return app.Synced(cmd.Context(), root, full)
}

func addReads(cmd *cobra.Command, root *string, output outputFunc) {
	var limit int
	var tag string
	ls := &cobra.Command{Use: "ls", Short: "List notes", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, args []string) error {
		if limit < 1 {
			return fmt.Errorf("limit must be positive")
		}
		idx, _, err := synced(cmd, *root, false)
		if err != nil {
			return err
		}
		defer idx.Close()
		result, err := idx.List(cmd.Context(), tag, limit)
		if err != nil {
			return err
		}
		var text strings.Builder
		for _, n := range result.Notes {
			fmt.Fprintf(&text, "%s\t%s\n", n.ID, n.Title)
		}
		return output(cmd, result, "%s", text.String())
	}}
	ls.Flags().IntVar(&limit, "limit", 50, "Maximum notes")
	ls.Flags().StringVar(&tag, "tag", "", "Filter by tag")
	cmd.AddCommand(ls)
	cmd.AddCommand(&cobra.Command{Use: "tags", Short: "List tags and counts", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, args []string) error {
		idx, _, err := synced(cmd, *root, false)
		if err != nil {
			return err
		}
		defer idx.Close()
		tags, err := idx.Tags(cmd.Context())
		if err != nil {
			return err
		}
		var text strings.Builder
		for _, t := range tags {
			fmt.Fprintf(&text, "%s\t%d\n", t.Name, t.Count)
		}
		return output(cmd, struct {
			Tags []index.Tag `json:"tags"`
		}{tags}, "%s", text.String())
	}})
	cmd.AddCommand(&cobra.Command{Use: "cat <id>", Short: "Read a note", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		idx, path, err := synced(cmd, *root, false)
		if err != nil {
			return err
		}
		defer idx.Close()
		n, raw, err := kb.Read(path, args[0])
		if err != nil {
			return err
		}
		result := struct {
			ID      string   `json:"id"`
			Title   string   `json:"title"`
			Tags    []string `json:"tags"`
			Aliases []string `json:"aliases"`
			Body    string   `json:"body"`
			Problem string   `json:"problem,omitempty"`
		}{n.ID, n.Title, n.Tags, n.Aliases, n.Body, n.Problem}
		return output(cmd, result, "%s", raw)
	}})
	var searchLimit int
	var keyword, semantic bool
	find := &cobra.Command{Use: "search <q>", Short: "Search notes by keyword or meaning", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		if searchLimit < 1 {
			return fmt.Errorf("limit must be positive")
		}
		idx, _, err := synced(cmd, *root, false)
		if err != nil {
			return err
		}
		defer idx.Close()
		mode := ""
		if keyword {
			mode = "keyword"
		}
		if semantic {
			mode = "semantic"
		}
		result, err := app.SearchDefault(cmd.Context(), idx, args[0], mode, searchLimit, cmd.ErrOrStderr())
		if err != nil {
			return err
		}
		var text strings.Builder
		for _, h := range result.Hits {
			fmt.Fprintf(&text, "%s\t%s\t%g\n%s\n", h.ID, h.Title, h.Score, h.Snippet)
		}
		return output(cmd, result, "%s", text.String())
	}}
	find.Flags().IntVar(&searchLimit, "limit", 50, "Maximum results")
	find.Flags().BoolVar(&keyword, "keyword", false, "Force keyword search")
	find.Flags().BoolVar(&semantic, "semantic", false, "Force semantic search")
	find.MarkFlagsMutuallyExclusive("keyword", "semantic")
	cmd.AddCommand(find)
	var full bool
	reindex := &cobra.Command{Use: "reindex", Short: "Synchronize the index", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, args []string) error {
		idx, _, err := synced(cmd, *root, full)
		if err != nil {
			return err
		}
		defer idx.Close()
		count, err := idx.Count(cmd.Context())
		if err != nil {
			return err
		}
		return output(cmd, struct {
			IndexedNotes int `json:"indexed_notes"`
		}{count}, "Indexed notes: %d\n", count)
	}}
	reindex.Flags().BoolVar(&full, "full", false, "Drop and rebuild the index")
	cmd.AddCommand(reindex)
}
