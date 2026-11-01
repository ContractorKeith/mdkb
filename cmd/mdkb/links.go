package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func addLinks(cmd *cobra.Command, root *string, output outputFunc) {
	for _, name := range []string{"links", "backlinks", "orphans", "broken"} {
		args := cobra.NoArgs
		use := name
		if name == "links" || name == "backlinks" {
			args = cobra.ExactArgs(1)
			use += " <id>"
		}
		cmd.AddCommand(&cobra.Command{Use: use, Short: "Report " + name, Args: args, RunE: func(cmd *cobra.Command, args []string) error {
			idx, _, err := synced(cmd, *root, false)
			if err != nil {
				return err
			}
			defer idx.Close()
			var result any
			var text strings.Builder
			switch name {
			case "links":
				r, err := idx.Links(cmd.Context(), args[0])
				if err != nil {
					return err
				}
				result = r
				for _, l := range r.Links {
					if l.Resolved {
						fmt.Fprintln(&text, *l.TargetID)
					} else {
						fmt.Fprintf(&text, "(broken) %s\n", l.Target)
					}
				}
			case "backlinks":
				r, err := idx.Backlinks(cmd.Context(), args[0])
				if err != nil {
					return err
				}
				result = r
				for _, l := range r.Backlinks {
					fmt.Fprintf(&text, "%s:%d\t%s\n", l.From, l.Line, l.Alias)
				}
			case "orphans":
				r, err := idx.Orphans(cmd.Context())
				if err != nil {
					return err
				}
				result = r
				for _, n := range r.Orphans {
					fmt.Fprintf(&text, "%s\t%s\n", n.ID, n.Title)
				}
			case "broken":
				r, err := idx.Broken(cmd.Context())
				if err != nil {
					return err
				}
				result = r
				for _, l := range r.Broken {
					fmt.Fprintf(&text, "%s:%d\t%s\t%s\n", l.From, l.Line, l.Target, l.Reason)
				}
			}
			return output(cmd, result, "%s", text.String())
		}})
	}
}
