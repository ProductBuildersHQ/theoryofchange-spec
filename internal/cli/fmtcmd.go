package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ProductBuildersHQ/impact-model/pkg/service"
)

func newFmtCmd(e *env) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fmt <file>...",
		Short: "Rewrite documents in canonical form",
		Long: `Rewrite documents in canonical form: stable field order, two-space
indentation, and a trailing newline, so diffs stay readable. Formatting is
idempotent. Element order is preserved; --sort orders each element list by
ID.

YAML comments are not preserved, because tool-written files are canonical.
Keep notes in description and narrative fields.

With --check, files are not written and the exit code is 1 if any file would
change.`,
		Example: `  impact-model fmt impact-model.yaml
  impact-model fmt --check *.yaml
  impact-model fmt --sort impact-model.yaml`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			format, err := outputFormat(cmd)
			if err != nil {
				return err
			}
			check, sortElements := flagBool(cmd, "check"), flagBool(cmd, "sort")
			results := make([]service.FormatResult, 0, len(args))
			anyChanged := false
			for _, path := range args {
				r, err := service.FormatFile(path, check, sortElements)
				if err != nil {
					return &ExitError{Code: ExitUsageErr, Err: err}
				}
				results = append(results, r)
				anyChanged = anyChanged || r.Changed
			}
			if format == "json" {
				if err := writeJSON(e.out, results); err != nil {
					return &ExitError{Code: ExitUsageErr, Err: err}
				}
			} else if !flagBool(cmd, "quiet") || check {
				for _, r := range results {
					switch {
					case r.Written:
						fmt.Fprintf(e.out, "formatted %s\n", r.Path)
					case r.Changed:
						fmt.Fprintf(e.out, "would reformat %s\n", r.Path)
					case !flagBool(cmd, "quiet"):
						fmt.Fprintf(e.out, "%s is already formatted\n", r.Path)
					}
				}
			}
			if check && anyChanged {
				return invalid()
			}
			return nil
		},
	}
	cmd.Flags().Bool("check", false, "do not write; exit 1 if any file would change")
	cmd.Flags().Bool("sort", false, "order each element list by ID")
	return cmd
}
