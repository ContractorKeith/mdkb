package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/ContractorKeith/mdkb/internal/config"
	"github.com/ContractorKeith/mdkb/internal/index"
	"github.com/ContractorKeith/mdkb/internal/kb"
	"github.com/spf13/cobra"
)

var version = "dev"

func main() {
	cmd := newRoot()
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "mdkb:", strings.Join(strings.Fields(err.Error()), " "))
		os.Exit(exitCode(err))
	}
}

func newRoot() *cobra.Command {
	var root string
	var jsonOutput, quiet bool
	cmd := &cobra.Command{
		Use: "mdkb", Short: "A local-first Markdown knowledge base",
		Version: version, SilenceUsage: true, SilenceErrors: true,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error { return cmd.Help() },
	}
	cmd.SetVersionTemplate("mdkb {{.Version}}\n")
	cmd.PersistentFlags().StringVar(&root, "root", "", "KB root (flag > MDKB_ROOT > config > ~/kb)")
	cmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Print JSON")
	cmd.PersistentFlags().BoolVar(&quiet, "quiet", false, "Suppress success output (except JSON)")
	output := func(cmd *cobra.Command, value any, format string, args ...any) error {
		if jsonOutput {
			return json.NewEncoder(cmd.OutOrStdout()).Encode(value)
		}
		if quiet {
			return nil
		}
		_, err := fmt.Fprintf(cmd.OutOrStdout(), format, args...)
		return err
	}
	cmd.AddCommand(&cobra.Command{
		Use: "version", Short: "Print the build version", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return output(cmd, struct {
				Version string `json:"version"`
			}{version}, "mdkb %s\n", version)
		},
	})
	var dryRun bool
	initCmd := &cobra.Command{
		Use: "init [path]", Short: "Initialize a KB in an empty directory", Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var path string
			var err error
			if len(args) == 1 {
				path, err = config.AbsolutePath(args[0])
			} else {
				path, err = config.ResolveRoot(root)
			}
			if err != nil {
				return err
			}
			if dryRun {
				err = kb.PreviewInit(path)
			} else {
				err = kb.Init(path)
			}
			if err != nil {
				return err
			}
			result := struct {
				Root   string `json:"root"`
				DryRun bool   `json:"dry_run"`
			}{path, dryRun}
			if dryRun {
				return output(cmd, result, "Would initialize KB: %s\n", path)
			}
			return output(cmd, result, "KB initialized: %s\n", path)
		},
	}
	initCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Validate initialization without writing")
	cmd.AddCommand(initCmd)
	cmd.AddCommand(&cobra.Command{
		Use: "info", Short: "Inspect the root and count Markdown notes", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := config.ResolveRoot(root)
			if err != nil {
				return err
			}
			info, err := kb.Info(path)
			if err != nil {
				return err
			}
			if info.Initialized {
				idx, err := index.Open(path)
				if err != nil {
					return err
				}
				defer idx.Close()
				if err := idx.Sync(cmd.Context(), path, false); err != nil {
					return err
				}
				info.IndexedNotes, err = idx.Count(cmd.Context())
				if err != nil {
					return err
				}
				info.IndexPath = idx.Path
			}
			return output(cmd, info, "Root: %s\nNotes: %d\nInitialized: %t\nIndexed notes: %d\nIndex: %s\nProblems: %d\n", info.Root, info.NoteCount, info.Initialized, info.IndexedNotes, info.IndexPath, info.Problems)
		},
	})
	addReads(cmd, &root, output)
	return cmd
}

// Map domain failures to the CLI contract; unclassified errors are usage/runtime errors.
func exitCode(err error) int {
	switch {
	case err == nil:
		return 0
	case errors.Is(err, errUninitialized):
		return 3
	case errors.Is(err, index.ErrIndex):
		return 4
	case errors.Is(err, kb.ErrWriteRefused):
		return 5
	case errors.Is(err, os.ErrNotExist):
		return 2
	default:
		return 1
	}
}
