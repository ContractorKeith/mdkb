package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ContractorKeith/mdkb/internal/config"
	"github.com/ContractorKeith/mdkb/internal/index"
	"github.com/ContractorKeith/mdkb/internal/kb"
	"github.com/ContractorKeith/mdkb/internal/search"
	"github.com/spf13/cobra"
)

var errUninitialized = errors.New("KB not initialized")

type outputFunc func(*cobra.Command, any, string, ...any) error

func synced(cmd *cobra.Command, root string, full bool) (*index.Index, string, error) {
	path, err := config.ResolveRoot(root)
	if err != nil {
		return nil, "", err
	}
	info, err := kb.Info(path)
	if err != nil || !info.Initialized {
		return nil, "", fmt.Errorf("%w: %s", errUninitialized, path)
	}
	idx, err := index.Open(path)
	if err != nil {
		return nil, "", err
	}
	if err := idx.Sync(cmd.Context(), path, full); err != nil {
		idx.Close()
		return nil, "", err
	}
	return idx, path, nil
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
		}{n.ID, n.Title, n.Tags, n.Aliases, n.Body}
		return output(cmd, result, "%s", raw)
	}})
	var searchLimit int
	find := &cobra.Command{Use: "search <q>", Short: "Search notes by keyword", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		if searchLimit < 1 {
			return fmt.Errorf("limit must be positive")
		}
		idx, _, err := synced(cmd, *root, false)
		if err != nil {
			return err
		}
		defer idx.Close()
		hits, err := (search.Search{Index: idx}).Keyword(cmd.Context(), args[0], searchLimit)
		if err != nil {
			return err
		}
		var text strings.Builder
		for _, h := range hits {
			fmt.Fprintf(&text, "%s\t%s\t%g\n%s\n", h.ID, h.Title, h.Score, h.Snippet)
		}
		return output(cmd, struct {
			Hits []search.Hit `json:"hits"`
		}{hits}, "%s", text.String())
	}}
	find.Flags().IntVar(&searchLimit, "limit", 50, "Maximum results")
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
