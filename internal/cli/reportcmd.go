package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/graph"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/report"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/service"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/tocio"
)

// render loads a document, produces text with fn, and writes it to --out or
// standard output. Documents the graph cannot be built from exit 1.
func render(e *env, cmd *cobra.Command, path string, fn func(*toc.Document) (string, error)) error {
	doc, err := tocio.LoadDocument(path)
	if err != nil {
		return &ExitError{Code: ExitUsageErr, Err: err}
	}
	text, err := fn(doc)
	if err != nil {
		if errors.Is(err, graph.ErrInvalidDocument) {
			return &ExitError{Code: ExitInvalid, Err: fmt.Errorf("%s: %w", path, err)}
		}
		return &ExitError{Code: ExitUsageErr, Err: err}
	}
	if out := flagString(cmd, "out"); out != "" {
		if err := tocio.WriteFile(out, []byte(text)); err != nil {
			return &ExitError{Code: ExitUsageErr, Err: err}
		}
		return nil
	}
	fmt.Fprint(e.out, text)
	return nil
}

func newReportCmd(e *env) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report <kind> <toc-file>",
		Short: "Render a report from a Theory of Change",
		Long: `Render a definition report. These need no observations, so a workshop can
review them before any data exists. Output is portable Markdown.

Kinds: ` + strings.Join(report.DefinitionKinds, ", ") + `.
"definitions" renders all of them as one document.`,
		Example: `  theoryofchange report definitions theory.yaml --out definitions.md
  theoryofchange report logic-model theory.yaml --pathway pw-supply
  theoryofchange report glossary theory.yaml`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return render(e, cmd, args[1], func(doc *toc.Document) (string, error) {
				return service.DefinitionReport(doc, args[0], flagString(cmd, "pathway"))
			})
		},
	}
	cmd.Flags().String("pathway", "", "restrict the report to one pathway ID")
	cmd.Flags().String("out", "", "write to this file instead of standard output")
	return cmd
}

func newExportCmd(e *env) *cobra.Command {
	export := &cobra.Command{
		Use:   "export",
		Short: "Export tables from a Theory of Change",
	}
	csv := &cobra.Command{
		Use:   "csv <table> <toc-file>",
		Short: "Export a table as CSV (RFC 4180)",
		Long: `Export a table for spreadsheets. Tables: ` + strings.Join(report.CSVTables, ", ") + `.
Multi-valued cells are joined with "; ".`,
		Example: `  theoryofchange export csv indicators theory.yaml --out indicators.csv`,
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return render(e, cmd, args[1], func(doc *toc.Document) (string, error) {
				return service.ExportCSV(doc, args[0])
			})
		},
	}
	csv.Flags().String("out", "", "write to this file instead of standard output")
	export.AddCommand(csv)
	return export
}
