package cli

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ProductBuildersHQ/impact-model/schema"
)

func newSchemaCmd(e *env) *cobra.Command {
	return &cobra.Command{
		Use:   "schema [name]",
		Short: "List the embedded JSON Schemas or print one",
		Long: `With no argument, lists the embedded schema names. With a name, prints that
JSON Schema, for editors and agents.`,
		Example: `  impact-model schema
  impact-model schema impact-model > impact-model.schema.json`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			format, err := outputFormat(cmd)
			if err != nil {
				return err
			}
			if len(args) == 1 {
				data, err := schema.Get(args[0])
				if err != nil {
					if errors.Is(err, schema.ErrUnknownSchema) {
						return usageErr("%v (available: %v)", err, schema.Names())
					}
					return &ExitError{Code: ExitUsageErr, Err: err}
				}
				if _, err := e.out.Write(data); err != nil {
					return &ExitError{Code: ExitUsageErr, Err: fmt.Errorf("write schema: %w", err)}
				}
				return nil
			}
			if format == "json" {
				if err := writeJSON(e.out, schema.Names()); err != nil {
					return &ExitError{Code: ExitUsageErr, Err: err}
				}
				return nil
			}
			for _, n := range schema.Names() {
				fmt.Fprintln(e.out, n)
			}
			return nil
		},
	}
}
