package cli

import (
	"context"
	"os"
	"os/signal"

	"github.com/spf13/cobra"

	"github.com/ProductBuildersHQ/impact-model/pkg/mcp"
)

func newMCPCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Serve the tools to an AI agent over MCP (stdio)",
		Long: `Start a Model Context Protocol server on standard input and output, so an AI
assistant can validate, lint, query, measure, report on, and edit an impact model.

Tools: impact_schema, impact_validate, impact_lint, impact_graph, impact_status, impact_report,
impact_patch, impact_deck, impact_judge_bundle, impact_diff. A tool takes an impact model inline
(document) or by path. A path must lie inside --root, which defaults to the
current directory: ".." is refused, and so is a symbolic link that leaves the
root. impact_patch rejects any patch that adds validation errors.

The server makes no network calls and no model calls.`,
		Example: `  impact-model mcp --root ./impact-models`,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			root := flagString(cmd, "root")
			srv, err := mcp.New(root)
			if err != nil {
				return &ExitError{Code: ExitUsageErr, Err: err}
			}
			parent := cmd.Context()
			if parent == nil {
				parent = context.Background()
			}
			ctx, stop := signal.NotifyContext(parent, os.Interrupt)
			defer stop()
			if err := srv.Run(ctx); err != nil && ctx.Err() == nil {
				return &ExitError{Code: ExitUsageErr, Err: err}
			}
			return nil
		},
	}
	cmd.Flags().String("root", ".", "the directory tool paths are confined to")
	return cmd
}
