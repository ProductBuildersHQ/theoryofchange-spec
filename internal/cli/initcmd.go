package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/service"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/tocio"
)

func newInitCmd(e *env) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Write a starter Theory of Change",
		Long: `Write a small starter Theory of Change that already passes validation, with
TODO placeholders to replace. The output is YAML unless --format json is
given or --out ends in .json.

Available templates: ` + strings.Join(service.Templates(), ", ") + `.`,
		Example: `  theoryofchange init --id youth-employment --name "Youth Employment" --out theory.yaml
  theoryofchange init --format json > theory.json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := flagString(cmd, "out")
			f, err := initFormat(cmd, out)
			if err != nil {
				return err
			}
			doc, err := service.InitDocument(service.InitOptions{
				Template: flagString(cmd, "template"),
				ID:       flagString(cmd, "id"),
				Name:     flagString(cmd, "name"),
				Problem:  flagString(cmd, "problem"),
			})
			if err != nil {
				return usageErr("%v", err)
			}
			data, err := tocio.Canonical(doc, f)
			if err != nil {
				return &ExitError{Code: ExitUsageErr, Err: err}
			}
			if out == "" {
				if _, err := e.out.Write(data); err != nil {
					return &ExitError{Code: ExitUsageErr, Err: fmt.Errorf("write output: %w", err)}
				}
				return nil
			}
			if _, err := os.Stat(out); err == nil && !flagBool(cmd, "force") {
				return usageErr("%s already exists; use --force to overwrite", out)
			} else if err != nil && !errors.Is(err, os.ErrNotExist) {
				return &ExitError{Code: ExitUsageErr, Err: fmt.Errorf("stat %s: %w", out, err)}
			}
			if err := tocio.WriteFile(out, data); err != nil {
				return &ExitError{Code: ExitUsageErr, Err: err}
			}
			if !flagBool(cmd, "quiet") {
				fmt.Fprintf(e.errOut, "wrote %s\n", out)
			}
			return nil
		},
	}
	// init's own --format selects the file format, not the report format.
	cmd.Flags().String("format", "", "file format: yaml or json (default: from --out, else yaml)")
	cmd.Flags().String("template", "minimal", "starter template")
	cmd.Flags().String("id", "", "document id (a lowercase kebab-case slug)")
	cmd.Flags().String("name", "", "document title")
	cmd.Flags().String("problem", "", "problem statement")
	cmd.Flags().String("out", "", "write to this file instead of stdout")
	cmd.Flags().Bool("force", false, "overwrite --out if it exists")
	return cmd
}

func initFormat(cmd *cobra.Command, out string) (tocio.Format, error) {
	switch flagString(cmd, "format") {
	case "json":
		return tocio.FormatJSON, nil
	case "yaml", "yml":
		return tocio.FormatYAML, nil
	case "":
		if out != "" {
			if f, err := tocio.FormatFromPath(out); err == nil {
				return f, nil
			}
		}
		return tocio.FormatYAML, nil
	default:
		return "", usageErr("--format must be yaml or json")
	}
}
