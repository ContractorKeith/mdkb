package main

import (
	"github.com/ContractorKeith/mdkb/internal/mcpserver"
	"github.com/spf13/cobra"
)

func addMCP(cmd *cobra.Command, root *string) {
	var allowWrite bool
	command := &cobra.Command{Use: "mcp", Short: "Serve MCP over stdio", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		return mcpserver.Run(cmd.Context(), *root, allowWrite)
	}}
	command.Flags().BoolVar(&allowWrite, "allow-write", false, "Register write_note and capture tools")
	cmd.AddCommand(command)
}
