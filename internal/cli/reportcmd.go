package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/graph"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/report"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/service"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/share"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/toc"
	"github.com/ProductBuildersHQ/theoryofchange-spec/pkg/tocio"
)

// asOfFlags adds the flags that control the export stamp.
func asOfFlags(cmd *cobra.Command) {
	cmd.Flags().String("as-of", "", "date the artifact describes (YYYY-MM-DD), recorded in the export stamp")
	cmd.Flags().Bool("no-stamp", false, "leave the export stamp off")
}

// stampFor builds the export stamp for doc from the --as-of flag.
func stampFor(cmd *cobra.Command, doc *toc.Document) (share.Stamp, error) {
	s, err := share.NewStamp(doc, flagString(cmd, "as-of"))
	if err != nil {
		return share.Stamp{}, &ExitError{Code: ExitUsageErr, Err: err}
	}
	return s, nil
}

// loadForRender loads a document and maps a failure to the exit code.
func loadForRender(path string) (*toc.Document, error) {
	doc, err := tocio.LoadDocument(path)
	if err != nil {
		return nil, &ExitError{Code: ExitUsageErr, Err: err}
	}
	return doc, nil
}

// renderError maps an error from building a view to an exit code.
func renderError(path string, err error) error {
	if errors.Is(err, graph.ErrInvalidDocument) {
		return &ExitError{Code: ExitInvalid, Err: fmt.Errorf("%s: %w", path, err)}
	}
	return &ExitError{Code: ExitUsageErr, Err: err}
}

// render loads a document, produces text with fn, stamps it (Markdown gets a
// metadata block, CSV a leading comment row) unless --no-stamp is set, and
// writes it to --out or standard output.
func render(e *env, cmd *cobra.Command, path string, csv bool, fn func(*toc.Document) (string, error)) error {
	doc, err := loadForRender(path)
	if err != nil {
		return err
	}
	text, err := fn(doc)
	if err != nil {
		return renderError(path, err)
	}
	if !flagBool(cmd, "no-stamp") {
		st, err := stampFor(cmd, doc)
		if err != nil {
			return err
		}
		if csv {
			text = share.StampCSV(st, text)
		} else {
			text = share.StampMarkdown(st, doc.Metadata.Name, text)
		}
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
			return render(e, cmd, args[1], false, func(doc *toc.Document) (string, error) {
				return service.DefinitionReport(doc, args[0], flagString(cmd, "pathway"))
			})
		},
	}
	cmd.Flags().String("pathway", "", "restrict the report to one pathway ID")
	cmd.Flags().String("out", "", "write to this file instead of standard output")
	asOfFlags(cmd)
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
Multi-valued cells are joined with "; ". The first row is a comment carrying the
export stamp ("# toc:<id> version:<v> hash:<h>"); --no-stamp leaves it off.`,
		Example: `  theoryofchange export csv indicators theory.yaml --out indicators.csv`,
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return render(e, cmd, args[1], true, func(doc *toc.Document) (string, error) {
				return service.ExportCSV(doc, args[0])
			})
		},
	}
	csv.Flags().String("out", "", "write to this file instead of standard output")
	asOfFlags(csv)
	export.AddCommand(csv, newExportDocCmd("docx"), newExportDocCmd("pptx"), newReviewPacketCmd(e), newReviewSheetCmd(e))
	return export
}
